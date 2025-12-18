package glacier

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/monitoring"
	goRedis "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// RestoreStatus represents the state of a GLACIER restore operation
type RestoreStatus struct {
	Key         string
	RequestedAt time.Time
	ExpiresAt   time.Time
	InProgress  bool
}

// RestoreCache tracks GLACIER restore operations to avoid duplicate requests
type RestoreCache interface {
	// MarkRestoreRequested records that a restore was initiated
	MarkRestoreRequested(ctx context.Context, key string, expiresIn time.Duration) error

	// GetRestoreStatus checks if restore was already requested
	GetRestoreStatus(ctx context.Context, key string) (*RestoreStatus, error)
}

// memoryRestoreCache implements RestoreCache using in-memory storage
type memoryRestoreCache struct {
	cache map[string]RestoreStatus
	mu    sync.RWMutex
}

// NewMemoryRestoreCache creates a new memory-based restore cache
func NewMemoryRestoreCache() RestoreCache {
	return &memoryRestoreCache{
		cache: make(map[string]RestoreStatus),
	}
}

func (m *memoryRestoreCache) MarkRestoreRequested(ctx context.Context, key string, expiresIn time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cache[key] = RestoreStatus{
		Key:         key,
		RequestedAt: time.Now(),
		ExpiresAt:   time.Now().Add(expiresIn),
		InProgress:  true,
	}

	return nil
}

func (m *memoryRestoreCache) GetRestoreStatus(ctx context.Context, key string) (*RestoreStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status, ok := m.cache[key]
	if !ok {
		return nil, nil // Not found
	}

	// Check if expired
	if time.Now().After(status.ExpiresAt) {
		return nil, nil // Expired, no longer in progress
	}

	return &status, nil
}

// redisRestoreCache implements RestoreCache using Redis
type redisRestoreCache struct {
	client goRedis.UniversalClient
}

// NewRedisRestoreCache creates a new Redis-based restore cache
func NewRedisRestoreCache(addresses []string) RestoreCache {
	var client goRedis.UniversalClient

	if len(addresses) > 1 {
		// Use cluster client for multiple addresses
		client = goRedis.NewClusterClient(&goRedis.ClusterOptions{
			Addrs: addresses,
		})
	} else {
		// Use single client
		client = goRedis.NewClient(&goRedis.Options{
			Addr: addresses[0],
		})
	}

	return &redisRestoreCache{client: client}
}

func (r *redisRestoreCache) getKey(key string) string {
	return fmt.Sprintf("mort:glacier:restore:%s", key)
}

func (r *redisRestoreCache) MarkRestoreRequested(ctx context.Context, key string, expiresIn time.Duration) error {
	cacheKey := r.getKey(key)

	// Store as "requestedAt:expiresAt" format
	data := fmt.Sprintf("%d:%d", time.Now().Unix(), time.Now().Add(expiresIn).Unix())

	// Set with TTL slightly longer than restore time to track completion
	return r.client.Set(ctx, cacheKey, data, expiresIn+24*time.Hour).Err()
}

func (r *redisRestoreCache) GetRestoreStatus(ctx context.Context, key string) (*RestoreStatus, error) {
	cacheKey := r.getKey(key)

	val, err := r.client.Get(ctx, cacheKey).Result()
	if err == goRedis.Nil {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, err
	}

	// Parse "requestedAt:expiresAt" format
	var reqAt, expAt int64
	if _, err := fmt.Sscanf(val, "%d:%d", &reqAt, &expAt); err != nil {
		return nil, fmt.Errorf("failed to parse restore status: %w", err)
	}

	status := &RestoreStatus{
		Key:         key,
		RequestedAt: time.Unix(reqAt, 0),
		ExpiresAt:   time.Unix(expAt, 0),
		InProgress:  time.Now().Before(time.Unix(expAt, 0)),
	}

	return status, nil
}

// CreateRestoreCache creates a RestoreCache based on configuration
func CreateRestoreCache(cacheCfg config.CacheCfg) RestoreCache {
	switch cacheCfg.Type {
	case "redis", "redis-cluster":
		monitoring.Log().Info("Creating Redis restore cache", zap.Strings("addresses", cacheCfg.Address))
		return NewRedisRestoreCache(cacheCfg.Address)
	default:
		monitoring.Log().Info("Creating memory restore cache")
		return NewMemoryRestoreCache()
	}
}

// Global singleton restore cache
var restoreCache RestoreCache
var restoreCacheOnce sync.Once

// GetRestoreCache returns the singleton restore cache instance
func GetRestoreCache(cacheCfg config.CacheCfg) RestoreCache {
	restoreCacheOnce.Do(func() {
		restoreCache = CreateRestoreCache(cacheCfg)
	})
	return restoreCache
}
