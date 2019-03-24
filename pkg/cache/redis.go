package cache

import (
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
		mp[parts[0]] = ":" + parts[1]
	}

	return mp
}

type RedisCache struct {
	client *redisCache.Codec
}

func NewRedis(redisAddress []string) *RedisCache {
	cache := redisCache.Codec{
		Redis: goRedis.NewRing(&goRedis.RingOptions{
			Addrs: parseAddress(redisAddress),
		}),

		Marshal: func(v interface{}) ([]byte, error) {
			return msgpack.Marshal(v)
		},
		Unmarshal: func(b []byte, v interface{}) error {
			return msgpack.Unmarshal(b, v)
		},
	}
	cache.UseLocalCache(10, time.Second*60)

	return &RedisCache{&cache}
}

func (c *RedisCache) Set(obj *object.FileObject, res *response.Response) error {
	item := redisCache.Item{
		Key:        obj.GetResponseCacheKey(),
		Object:     res,
		Expiration: time.Second * time.Duration(res.GetTTL()),
	}
	return c.client.Set(&item)
}

func (c *RedisCache) Get(obj *object.FileObject) (*response.Response, error) {
	var res response.Response
	err := c.client.Get(obj.GetResponseCacheKey(), &res)
	return &res, err
}

func (c *RedisCache) Delete(obj *object.FileObject) error {
	return c.client.Delete(obj.GetResponseCacheKey())
}
