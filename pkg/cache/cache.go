package cache

import (
	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/monitoring"
	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/response"
	"go.uber.org/zap"
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
		monitoring.Log().Info("Creating redis cache", zap.Strings("addr", cacheCfg.Address))
		return NewRedis(cacheCfg.Address, cacheCfg.ClientConfig)
	case "redis-cluster":
		monitoring.Log().Info("Creating redis-cluster cache", zap.Strings("addr", cacheCfg.Address))
		return NewRedisCluster(cacheCfg.Address, cacheCfg.ClientConfig)
	default:
		monitoring.Log().Info("Creating memory cache")
		return NewMemoryCache(cacheCfg.CacheSize)
	}
}
