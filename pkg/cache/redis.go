package cache

import (
	"context"
	"strings"
	"time"

	"github.com/aldor007/mort/pkg/monitoring"
	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/response"
	redisCache "github.com/go-redis/cache/v8"
	goRedis "github.com/go-redis/redis/v8"
	"github.com/vmihailenco/msgpack"
)

func parseAddress(addrs []string) map[string]string {
	mp := make(map[string]string, len(addrs))

	for _, addr := range addrs {
		parts := strings.Split(addr, ":")
		mp[parts[0]] = parts[0] + ":" + parts[1]
	}

	return mp
}

type CacheCfg struct {
	MaxItemSize int64
	MinUseCount uint64
}

type redisClient interface {
	Incr(ctx context.Context, key string) *goRedis.IntCmd
	Get(ctx context.Context, key string) *goRedis.StringCmd
	Del(ctx context.Context, keys ...string) *goRedis.IntCmd
}

// RedisCache store response in redis
type RedisCache struct {
	cache  *redisCache.Cache
	client redisClient

	cfg CacheCfg
}

// NewRedis create connection to redis and update it config from clientConfig map
func NewRedis(redisAddress []string, clientConfig map[string]string, cfg CacheCfg) *RedisCache {
	ring := goRedis.NewRing(&goRedis.RingOptions{
		Addrs: parseAddress(redisAddress),
	})

	cache := redisCache.New(&redisCache.Options{
		Redis:      ring,
		LocalCache: redisCache.NewTinyLFU(10, time.Minute),
	})

	if clientConfig != nil {
		for key, value := range clientConfig {
			ring.ConfigSet(context.Background(), key, value)
		}
	}

	return &RedisCache{cache, ring, cfg}
}

func NewRedisCluster(redisAddress []string, clientConfig map[string]string, cfg CacheCfg) *RedisCache {
	ring := goRedis.NewClusterClient(&goRedis.ClusterOptions{
		Addrs: redisAddress,
	})
	cache := redisCache.New(&redisCache.Options{
		Redis:      ring,
		LocalCache: redisCache.NewTinyLFU(10, time.Minute),
	})

	if clientConfig != nil {
		for key, value := range clientConfig {
			ring.ConfigSet(context.Background(), key, value)
		}
	}

	return &RedisCache{cache, ring, cfg}
}

func (c *RedisCache) getKey(obj *object.FileObject) string {
	return "mort-v1:" + obj.GetResponseCacheKey()
}

// Set put response into cache
func (c *RedisCache) Set(obj *object.FileObject, res *response.Response) error {
	if res.ContentLength > c.cfg.MaxItemSize {
		return nil
	}

	if c.cfg.MinUseCount > 0 {
		countKey := "count" + c.getKey(obj)
		r := c.client.Incr(obj.Ctx, countKey)
		if counter, err := r.Uint64(); err != nil && counter < c.cfg.MinUseCount {
			return nil
		}
		c.client.Del(obj.Ctx, countKey)
	}

	monitoring.Report().Inc("cache_ratio;status:set")
	v, err := msgpack.Marshal(res)
	if err != nil {
		return err
	}
	item := redisCache.Item{
		Key:   c.getKey(obj),
		Value: v,
		TTL:   time.Second * time.Duration(res.GetTTL()),
	}
	return c.cache.Set(obj.Ctx, &item)
}

// Get returns response from cache or error
func (c *RedisCache) Get(obj *object.FileObject) (*response.Response, error) {
	var buf []byte
	var res response.Response
	err := c.cache.Get(obj.Ctx, c.getKey(obj), &buf)
	if err != nil {
		monitoring.Report().Inc("cache_ratio;status:miss")
	} else {
		monitoring.Report().Inc("cache_ratio;status:hit")
		err = msgpack.Unmarshal(buf, &res)
		if res.Headers != nil {
			res.SetCacheHit()
		}
	}

	return &res, err
}

// Delete remove response from cache
func (c *RedisCache) Delete(obj *object.FileObject) error {
	return c.cache.Delete(obj.Ctx, c.getKey(obj))
}
