package cache

import (
	"math"
	"time"
	"unsafe"

	"github.com/aldor007/mort/pkg/monitoring"
	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/response"
	"github.com/karlseguin/ccache"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type (
	// MemoryCache uses memory for cache purpose
	MemoryCache struct {
		cache *ccache.Cache // cache for created image transformations
	}

	// responseSizeProvider adapts response.Response to how ccache size computation requirements.
	responseSizeProvider struct {
		*response.Response
	}
)

// Size provides ths size of cached response in an naive way
func (r responseSizeProvider) Size() int64 {
	body, err := r.Response.Body()
	if err != nil {
		// Result from Size() method is counted to an overall cache limit.
		// Thus if the size cannot be determined it make sense to return MaxInt value instead of 0.
		return math.MaxInt64
	}
	size := len(body) + int(unsafe.Sizeof(*r.Response)) + int(unsafe.Sizeof(r.Response.Headers)) // map structures
	for k, v := range r.Response.Headers {
		for i := 0; i < len(v); i++ {
			size += len(v[i])
		}
		size += len(k)
	}
	// 350 bytes of overhead described in ccache documentation
	return int64(size) + 350
}

// NewMemoryCache returns instance of memory cache
func NewMemoryCache(maxSize int64) *MemoryCache {
	return &MemoryCache{ccache.New(ccache.Configure().MaxSize(maxSize).ItemsToPrune(50))}
}

// Set put response to cache
func (c *MemoryCache) Set(obj *object.FileObject, res *response.Response) error {
	cachedResp, err := res.Copy()
	if err != nil {
		return err
	}
	monitoring.Report().Inc("cache_ratio;status:set")
	c.cache.Set(obj.GetResponseCacheKey(), responseSizeProvider{cachedResp}, time.Second*time.Duration(res.GetTTL()))
	return nil
}

// Get returns instance from cache or error (if not found in cache)
func (c *MemoryCache) Get(obj *object.FileObject) (*response.Response, error) {
	cacheValue := c.cache.Get(obj.GetResponseCacheKey())
	if cacheValue != nil {
		monitoring.Log().Info("Handle Get cache", zap.String("cache", "hit"), zap.String("obj.Key", obj.Key))
		monitoring.Report().Inc("cache_ratio;status:hit")
		res := cacheValue.Value().(responseSizeProvider)
		resCp, err := res.Copy()
		if err != nil {
			monitoring.Report().Inc("cache_ratio;status:miss")
			return nil, errors.New("not found")
		}
		resCp.Set("x-mort-cache", "hit")
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
