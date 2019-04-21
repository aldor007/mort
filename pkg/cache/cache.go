package cache

import (
	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/response"
)


// ResponseCache is interface for caching of mort responses
type ResponseCache interface {
	Set(obj *object.FileObject, res *response.Response) error
	Get(obj *object.FileObject) (*response.Response, error)
	Delete(obj *object.FileObject) error
}


// Create returns instance of Response cache
func Create(cacheCfg config.CacheCfg) ResponseCache {
	switch cacheCfg.Type {
	case "redis":
		return NewRedis(cacheCfg.Address, cacheCfg.ClientConfig)
	default:
		return NewMemoryCache(cacheCfg.CacheSize)
	}
}
