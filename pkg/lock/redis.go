package lock

import (
	"sync"
	"time"

	"github.com/aldor007/mort/pkg/response"
	"github.com/bsm/redislock"
	goRedis "github.com/go-redis/redis/v8"

	"context"
	"strings"
)

func parseAddress(addrs []string) map[string]string {
	mp := make(map[string]string, len(addrs))

	for _, addr := range addrs {
		parts := strings.Split(addr, ":")
		mp[parts[0]] = parts[0] + ":" + parts[1]
	}

	return mp
}

// RedisLock is in Redis lock for single mort instance
type RedisLock struct {
	client     *redislock.Client
	memoryLock *MemoryLock
	locks      map[string]*redislock.Lock
	lock       sync.RWMutex
}

// NewRedis create connection to redis and update it config from clientConfig map
func NewRedisLock(redisAddress []string, clientConfig map[string]string) *RedisLock {
	ring := goRedis.NewRing(&goRedis.RingOptions{
		Addrs: parseAddress(redisAddress),
	})

	if clientConfig != nil {
		for key, value := range clientConfig {
			ring.ConfigSet(context.Background(), key, value)
		}
	}

	locker := redislock.New(ring)

	return &RedisLock{client: locker, memoryLock: NewMemoryLock(), locks: make(map[string]*redislock.Lock)}
}

func NewRedisCluster(redisAddress []string, clientConfig map[string]string) *RedisLock {
	ring := goRedis.NewClusterClient(&goRedis.ClusterOptions{
		Addrs: redisAddress,
	})
	if clientConfig != nil {
		for key, value := range clientConfig {
			ring.ConfigSet(context.Background(), key, value)
		}
	}

	locker := redislock.New(ring)

	return &RedisLock{client: locker, memoryLock: NewMemoryLock(), locks: make(map[string]*redislock.Lock)}
}

// NotifyAndRelease tries notify all waiting goroutines about response
func (m *RedisLock) NotifyAndRelease(ctx context.Context, key string, originalResponse *response.Response) {
	m.lock.Lock()
	defer m.lock.Unlock()
	lock, ok := m.locks[key]
	if ok {
		lock.Release(ctx)
		delete(m.locks, key)
	}

	m.memoryLock.NotifyAndRelease(ctx, key, originalResponse)
}

// Lock create unique entry in Redis map
func (m *RedisLock) Lock(ctx context.Context, key string) (result LockResult, ok bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	lock, ok := m.locks[key]
	if ok {
		lock.Refresh(ctx, time.Millisecond*500, nil)
	} else {
		lock, err := m.client.Obtain(ctx, key, 60*time.Second, nil)
		if err != nil {
			result.Error = err
			return
		}
		m.locks[key] = lock

	}
	return m.memoryLock.Lock(ctx, key)
}

// Release remove entry from Redis map
func (m *RedisLock) Release(ctx context.Context, key string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	lock, ok := m.locks[key]
	if ok {
		lock.Release(ctx)
		delete(m.locks, key)
		m.memoryLock.Release(ctx, key)
	}
}
