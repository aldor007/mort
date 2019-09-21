package cache

import (
	"github.com/aldor007/mort/pkg/monitoring"
	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/response"
	redisCache "github.com/go-redis/cache"
	goRedis "github.com/go-redis/redis"
	"github.com/vmihailenco/msgpack"
	"strings"
	"time"
)

func parseAddress(addrs []string) map[string]string {
	mp := make(map[string]string, len(addrs))

	for _, addr := range addrs {
		parts := strings.Split(addr, ":")
		mp[parts[0]] = parts[0] + ":" + parts[1]
	}

	return mp
}

// RedisCache store response in redis
type RedisCache struct {
	client *redisCache.Codec
}

// NewRedis create connection to redis and update it config from clientConfig map
func NewRedis(redisAddress []string, clientConfig map[string]string) *RedisCache {
	ring := goRedis.NewRing(&goRedis.RingOptions{
		Addrs: parseAddress(redisAddress),
	})
	cache := redisCache.Codec{
		Redis: ring,

		Marshal: func(v interface{}) ([]byte, error) {
			return msgpack.Marshal(v)
		},
		Unmarshal: func(b []byte, v interface{}) error {
			return msgpack.Unmarshal(b, v)
		},
	}
	cache.UseLocalCache(10, time.Second*60)
	if clientConfig != nil {
		for key, value := range clientConfig {
			ring.ConfigSet(key, value)
		}
	}

	return &RedisCache{&cache}
}

// Set put response into cache
func (c *RedisCache) Set(obj *object.FileObject, res *response.Response) error {
	monitoring.Report().Inc("cache_ratio;status:set")

	item := redisCache.Item{
		Key:        obj.GetResponseCacheKey(),
		Object:     res,
		Expiration: time.Second * time.Duration(res.GetTTL()),
	}
	return c.client.Set(&item)
}

// Get returns response from cache or error
func (c *RedisCache) Get(obj *object.FileObject) (*response.Response, error) {
	var res response.Response
	err := c.client.Get(obj.GetResponseCacheKey(), &res)
	if err != nil {
		monitoring.Report().Inc("cache_ratio;status:miss")
	} else {
		monitoring.Report().Inc("cache_ratio;status:hit")
		res.Set("x-mort-cache", "hit")
	}

	return &res, err
}

// Delete remove response from cache
func (c *RedisCache) Delete(obj *object.FileObject) error {
	return c.client.Delete(obj.GetResponseCacheKey())
}
