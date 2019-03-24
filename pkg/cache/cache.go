package cache

import (
	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/response"
)

type ResponseCache interface {
	Set(obj *object.FileObject, res *response.Response) error
	Get(obj *object.FileObject) (*response.Response, error)
	Delete(obj *object.FileObject) error
}

func Create(cacheCfg config.CacheCfg) ResponseCache {
	switch cacheCfg.Type {
	case "redis":
		return NewRedis(cacheCfg.Address)
	default:
		return NewMemoryCache(cacheCfg.CacheSize)
	}
}
