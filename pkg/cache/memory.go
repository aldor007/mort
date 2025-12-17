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

// Set put response to cache. Cache takes ownership of the response - no copying.
// The response must be buffered before caching. This eliminates one full copy
// compared to the previous implementation that copied on both Set and Get.
func (c *MemoryCache) Set(obj *object.FileObject, res *response.Response) error {
	// Ensure response is buffered before caching
	if !res.IsBuffered() {
		_, err := res.Body()
		if err != nil {
			return err
		}
	}

	monitoring.Report().Inc("cache_ratio;status:set")

	// Cache takes ownership - NO COPY!
	// Calculate size once when creating cache entry
	provider := responseSizeProvider{
		Response:   res,
		cachedSize: calculateResponseSize(res),
	}
	c.cache.Set(obj.GetResponseCacheKey(), provider, time.Second*time.Duration(res.GetTTL()))
	return nil
}

// Get returns a view of the cached response (zero-copy).
// The view shares the underlying buffer with the cached response, eliminating
// the need to copy the full response body on every cache hit.
func (c *MemoryCache) Get(obj *object.FileObject) (*response.Response, error) {
	cacheValue := c.cache.Get(obj.GetResponseCacheKey())
	if cacheValue != nil {
		monitoring.Log().Info("Handle Get cache", zap.String("cache", "hit"), zap.String("obj.Key", obj.Key))
		monitoring.Report().Inc("cache_ratio;status:hit")

		cached := cacheValue.Value().Response

		// Create view instead of copy - zero memory allocation for body!
		view, err := cached.CreateView()
		if err != nil {
			monitoring.Report().Inc("cache_ratio;status:miss")
			return nil, errors.New("not found")
		}
		view.SetCacheHit()
		return view, nil
	}

	monitoring.Report().Inc("cache_ratio;status:miss")
	return nil, errors.New("not found")
}

// Delete remove given response from cache
func (c *MemoryCache) Delete(obj *object.FileObject) error {
	c.cache.Delete(obj.GetResponseCacheKey())
	return nil
}
