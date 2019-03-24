
package cache

import (
	"github.com/aldor007/mort/pkg/monitoring"
	"github.com/aldor007/mort/pkg/object"
"github.com/aldor007/mort/pkg/response"
	"github.com/karlseguin/ccache"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"
)

type MemoryCache struct {
	cache          *ccache.Cache          // cache for created image transformations
}

func NewMemoryCache(maxSize int64) *MemoryCache {
	return &MemoryCache{ccache.New(ccache.Configure().MaxSize(maxSize))}
}

func (c *MemoryCache) Set(obj *object.FileObject, res *response.Response) error {
	c.cache.Set(obj.GetResponseCacheKey(), res, time.Second * time.Duration(res.GetTTL()))
	return nil
}

func (c *MemoryCache) Get(obj *object.FileObject) (*response.Response, error) {
	cacheValue := c.cache.Get(obj.GetResponseCacheKey())
	if cacheValue != nil {
		monitoring.Log().Info("Handle Get cache", zap.String("cache", "hit"), zap.String("obj.Key", obj.Key))
		monitoring.Report().Inc("cache_ratio;status:hit")
		res := cacheValue.Value().(*response.Response)
		resCp, err := res.Copy()
		if err == nil {
			return resCp, nil
		}
	}

	return nil, errors.New("not found")
}

func (c *MemoryCache) Delete(obj *object.FileObject) error {
	c.cache.Delete(obj.GetResponseCacheKey())
	return nil
}
