package cache

import (
	"context"
	"sync"
	"time"

	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/monitoring"
	goRedis "github.com/go-redis/redis/v8"
	"github.com/vmihailenco/msgpack"
	"go.uber.org/zap"
)

// Cacheable represents types that can be cached
// Types must be msgpack-serializable (have exported fields or be basic types)
type Cacheable interface {
	any
}

// Cache is a generic cache interface for storing msgpack-serializable types
// T must have exported fields for proper serialization
// Examples: structs with exported fields, basic types (string, int, bool), slices, maps
type Cache[T Cacheable] interface {
	Set(ctx context.Context, key string, value T, ttl time.Duration) error
	Get(ctx context.Context, key string) (T, bool, error)
	Delete(ctx context.Context, key string) error
}

// GenericMemoryCache is a generic in-memory cache implementation
type GenericMemoryCache[T Cacheable] struct {
	cache map[string]cacheEntry[T]
	mu    sync.RWMutex
}

type cacheEntry[T Cacheable] struct {
	value     T
	expiresAt time.Time
}

// NewGenericMemoryCache creates a new generic memory cache
func NewGenericMemoryCache[T Cacheable]() Cache[T] {
	return &GenericMemoryCache[T]{
		cache: make(map[string]cacheEntry[T]),
	}
}

func (m *GenericMemoryCache[T]) Set(ctx context.Context, key string, value T, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	m.cache[key] = cacheEntry[T]{
		value:     value,
		expiresAt: expiresAt,
	}

	return nil
}

func (m *GenericMemoryCache[T]) Get(ctx context.Context, key string) (T, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var zero T
	entry, ok := m.cache[key]
	if !ok {
		return zero, false, nil
	}

	// Check expiration
	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		return zero, false, nil
	}

	return entry.value, true, nil
}

func (m *GenericMemoryCache[T]) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.cache, key)
	return nil
}

// GenericRedisCache is a generic Redis cache implementation using msgpack serialization
type GenericRedisCache[T Cacheable] struct {
	client goRedis.UniversalClient
}

// NewGenericRedisCache creates a new generic Redis cache using shared connection pool
func NewGenericRedisCache[T Cacheable](addresses []string, clientConfig map[string]string, cluster bool) Cache[T] {
	// Get shared Redis client from pool (reused across all Cache[T] instances)
	client := getRedisClient(addresses, clientConfig, cluster)

	return &GenericRedisCache[T]{client: client}
}

func (r *GenericRedisCache[T]) Set(ctx context.Context, key string, value T, ttl time.Duration) error {
	// Serialize value using msgpack (more efficient than JSON)
	data, err := msgpack.Marshal(value)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *GenericRedisCache[T]) Get(ctx context.Context, key string) (T, bool, error) {
	var zero T

	val, err := r.client.Get(ctx, key).Result()
	if err == goRedis.Nil {
		return zero, false, nil
	}
	if err != nil {
		return zero, false, err
	}

	// Deserialize from msgpack
	var result T
	if err := msgpack.Unmarshal([]byte(val), &result); err != nil {
		return zero, false, err
	}

	return result, true, nil
}

func (r *GenericRedisCache[T]) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// CreateCache creates a generic cache based on configuration
// T must be msgpack-serializable (structs with exported fields, basic types, etc.)
func CreateCache[T Cacheable](cacheCfg config.CacheCfg) Cache[T] {
	switch cacheCfg.Type {
	case "redis", "redis-cluster":
		monitoring.Log().Info("Creating generic cache (Redis with msgpack)", zap.Strings("addr", cacheCfg.Address))
		return NewGenericRedisCache[T](cacheCfg.Address, cacheCfg.ClientConfig, cacheCfg.Type == "redis-cluster")
	default:
		monitoring.Log().Info("Creating generic cache (Memory)")
		return NewGenericMemoryCache[T]()
	}
}
