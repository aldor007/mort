package cache

import (
	"math"
	"time"
	"unsafe"

	"github.com/aldor007/mort/pkg/monitoring"
	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/response"
	"github.com/karlseguin/ccache/v3"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type (
	// MemoryCache uses memory for cache purpose
	MemoryCache struct {
		cache *ccache.Cache[responseSizeProvider] // cache for created image transformations
	}

	// responseSizeProvider adapts response.Response to how ccache size computation requirements.
	responseSizeProvider struct {
		*response.Response
		cachedSize int64 // Pre-calculated size to avoid recalculation
	}
)

// Size returns the pre-calculated size of the cached response
func (r responseSizeProvider) Size() int64 {
	return r.cachedSize
}

// calculateResponseSize computes the size once during cache entry creation
func calculateResponseSize(res *response.Response) int64 {
	// Use ContentLength if available (more accurate and faster)
	size := res.ContentLength
	if size <= 0 {
		// If not available, try to get body size if already buffered
		if res.IsBuffered() {
			body, err := res.Body()
			if err != nil {
				// Return large value to prevent caching problematic responses
				return math.MaxInt64
			}
			size = int64(len(body))
		} else {
			// For unbuffered responses, use a conservative estimate
			size = 1024 * 1024 // 1MB default estimate
		}
	}

	// Add header overhead (estimate)
	headerSize := int64(unsafe.Sizeof(res.Headers))
	for k, v := range res.Headers {
		headerSize += int64(len(k))
		for i := 0; i < len(v); i++ {
			headerSize += int64(len(v[i]))
		}
	}

	// Add ccache overhead (350 bytes) + response struct overhead
	return size + headerSize + 350 + int64(unsafe.Sizeof(*res))
}

// NewMemoryCache returns instance of memory cache
func NewMemoryCache(maxSize int64) *MemoryCache {
	return &MemoryCache{ccache.New[responseSizeProvider](ccache.Configure[responseSizeProvider]().MaxSize(maxSize).ItemsToPrune(50))}
}

// Set put response to cache
func (c *MemoryCache) Set(obj *object.FileObject, res *response.Response) error {
	cachedResp, err := res.Copy()
	if err != nil {
		return err
	}
	monitoring.Report().Inc("cache_ratio;status:set")
	// Calculate size once when creating cache entry
	provider := responseSizeProvider{
		Response:   cachedResp,
		cachedSize: calculateResponseSize(cachedResp),
	}
	c.cache.Set(obj.GetResponseCacheKey(), provider, time.Second*time.Duration(res.GetTTL()))
	return nil
}

// Get returns instance from cache or error (if not found in cache)
func (c *MemoryCache) Get(obj *object.FileObject) (*response.Response, error) {
	cacheValue := c.cache.Get(obj.GetResponseCacheKey())
	if cacheValue != nil {
		monitoring.Log().Info("Handle Get cache", zap.String("cache", "hit"), zap.String("obj.Key", obj.Key))
		monitoring.Report().Inc("cache_ratio;status:hit")
		res := cacheValue.Value()
		resCp, err := res.Copy()
		if err != nil {
			monitoring.Report().Inc("cache_ratio;status:miss")
			return nil, errors.New("not found")
		}
		resCp.SetCacheHit()
		return resCp, nil
	}

	monitoring.Report().Inc("cache_ratio;status:miss")
	return nil, errors.New("not found")
}

// Delete remove given response from cache
func (c *MemoryCache) Delete(obj *object.FileObject) error {
	c.cache.Delete(obj.GetResponseCacheKey())
	return nil
}
