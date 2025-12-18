package lock

import (
	"github.com/aldor007/mort/pkg/monitoring"
	"go.uber.org/zap"
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

type rediser interface {
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *goRedis.BoolCmd
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) *goRedis.Cmd
	EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) *goRedis.Cmd
	ScriptExists(ctx context.Context, scripts ...string) *goRedis.BoolSliceCmd
	ScriptLoad(ctx context.Context, script string) *goRedis.StringCmd
	Subscribe(ctx context.Context, channels ...string) *goRedis.PubSub
	Publish(ctx context.Context, channel string, message interface{}) *goRedis.IntCmd
	Close() error
}

type internalLockRedis struct {
	lockData
	lock          *redislock.Lock
	pubsub        *goRedis.PubSub
	cancelRefresh context.CancelFunc // cancels the background refresh goroutine
}

// RedisLock is in Redis lock for single mort instance
type RedisLock struct {
	client      *redislock.Client
	memoryLock  *MemoryLock
	locks       map[string]internalLockRedis
	lock        sync.RWMutex
	LockTimeout int
	redisClient rediser
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

	return &RedisLock{client: locker, memoryLock: NewMemoryLock(), locks: make(map[string]internalLockRedis), LockTimeout: 60, redisClient: ring}
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

	return &RedisLock{client: locker, memoryLock: NewMemoryLock(), locks: make(map[string]internalLockRedis), LockTimeout: 60, redisClient: ring}
}

// startLockRefresh starts a background goroutine that periodically refreshes the Redis lock
// to prevent it from expiring during long-running operations. Returns a cancel function
// that must be called to stop the refresh goroutine and prevent goroutine leaks.
func (m *RedisLock) startLockRefresh(key string, lock *redislock.Lock) context.CancelFunc {
	refreshCtx, cancel := context.WithCancel(context.Background())
	refreshInterval := time.Duration(m.LockTimeout/2) * time.Second

	go func() {
		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()

		for {
			select {
			case <-refreshCtx.Done():
				return
			case <-ticker.C:
				err := lock.Refresh(refreshCtx, time.Duration(m.LockTimeout)*time.Second, nil)
				if err != nil {
					if err == context.Canceled {
						// Context was cancelled, normal shutdown
						return
					}
					monitoring.Log().Warn("Redis lock auto-refresh failed", zap.String("key", key), zap.Error(err))
				} else {
					monitoring.Log().Debug("Redis lock auto-refreshed", zap.String("key", key))
				}
			}
		}
	}()

	return cancel
}

// NotifyAndRelease tries notify all waiting goroutines about response
// Delegates to MemoryLock which uses SharedResponse for zero-copy buffer sharing
func (m *RedisLock) NotifyAndRelease(ctx context.Context, key string, sharedResponse *response.SharedResponse) {
	m.lock.Lock()
	defer m.lock.Unlock()
	lock, ok := m.locks[key]
	if ok {
		// Stop the background refresh goroutine before releasing the lock
		if lock.cancelRefresh != nil {
			lock.cancelRefresh()
		}

		if lock.lock != nil {
			err := lock.lock.Release(ctx)
			if err != nil {
				// Only log errors other than "lock not held" which can happen
				// if lock expired despite refresh attempts (e.g., network issues)
				if err != redislock.ErrNotObtained {
					monitoring.Log().Error("redis release error", zap.String("key", key), zap.Error(err))
				}
			}
		}
		delete(m.locks, key)
		if lock.pubsub != nil {
			lock.pubsub.Close()
		} else if lock.lock != nil {
			m.redisClient.Publish(ctx, key, 1)
		}
	}

	// Delegate to memory lock which handles SharedResponse reference counting
	m.memoryLock.NotifyAndRelease(ctx, key, sharedResponse)
}

// Lock create unique entry in Redis map
func (m *RedisLock) Lock(ctx context.Context, key string) (LockResult, bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	lock, ok := m.locks[key]
	if ok {
		if lock.lock != nil {
			// Use background context for refresh to ensure it works even if request context is cancelled
			refreshCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err := lock.lock.Refresh(refreshCtx, time.Second*time.Duration(m.LockTimeout/2), nil)
			cancel()
			if err != nil {
				monitoring.Log().Error("Redis lock refresh err", zap.String("key", key), zap.Error(err))
			}
		}
	} else {
		lock, err := m.client.Obtain(ctx, key, time.Duration(m.LockTimeout)*time.Second, nil)

		ok = false
		if err == redislock.ErrNotObtained {
			result, _ := m.memoryLock.forceLockAndAddWatch(ctx, key)
			pubsub := m.redisClient.Subscribe(ctx, key)
			lockData := internalLockRedis{lock: lock, pubsub: pubsub}
			// Go channel which receives messages.
			ch := pubsub.Channel()

			go func() {
				for {
					select {
					case <-ch:
						m.NotifyAndRelease(ctx, key, nil)
						return
					case <-result.Cancel:
						m.memoryLock.Release(ctx, key)
						err := pubsub.Close()
						if err != nil {
							monitoring.Log().Error("Redis lock pubsub err", zap.String("key", key), zap.Error(err))
						}
						return
					case <-ctx.Done():
						m.memoryLock.Release(ctx, key)
						err := pubsub.Close()
						if err != nil {
							monitoring.Log().Error("Redis lock pubsub err", zap.String("key", key), zap.Error(err))
						}
						return
					}
				}
			}()
			m.locks[key] = lockData
		} else if err != nil {
			return LockResult{Error: err}, false
		} else {
			// Successfully obtained lock, start background refresh
			cancelRefresh := m.startLockRefresh(key, lock)
			lockData := internalLockRedis{lock: lock, pubsub: nil, cancelRefresh: cancelRefresh}
			m.locks[key] = lockData
		}
	}
	return m.memoryLock.Lock(ctx, key)
}

// Release remove entry from Redis map
func (m *RedisLock) Release(ctx context.Context, key string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	lock, ok := m.locks[key]
	if ok {
		// Stop the background refresh goroutine before releasing the lock
		if lock.cancelRefresh != nil {
			lock.cancelRefresh()
		}

		if lock.lock != nil {
			lock.lock.Release(ctx)
		}
		if lock.pubsub != nil {
			lock.pubsub.Close()
		}
		delete(m.locks, key)
		m.memoryLock.Release(ctx, key)
	}
}

// Close closes the Redis client and cleans up all resources.
// This should be called when the RedisLock is no longer needed to prevent goroutine leaks.
func (m *RedisLock) Close() error {
	m.lock.Lock()
	defer m.lock.Unlock()

	// Clean up all active locks
	for key, lock := range m.locks {
		if lock.cancelRefresh != nil {
			lock.cancelRefresh()
		}
		if lock.lock != nil {
			lock.lock.Release(context.Background())
		}
		if lock.pubsub != nil {
			lock.pubsub.Close()
		}
		delete(m.locks, key)
	}

	// Close the Redis client to stop its background goroutines
	if m.redisClient != nil {
		return m.redisClient.Close()
	}
	return nil
}
