package storage

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aldor007/mort/pkg/cache"
	"github.com/aldor007/mort/pkg/config"
)

// RestoreStatus represents the state of a GLACIER restore operation
// All fields are exported for msgpack serialization
type RestoreStatus struct {
	Key         string    `msgpack:"key"`
	RequestedAt time.Time `msgpack:"requestedAt"`
	ExpiresAt   time.Time `msgpack:"expiresAt"`
	InProgress  bool      `msgpack:"inProgress"`
}

// RestoreCache handles GLACIER restore status tracking using generic cache
type RestoreCache struct {
	cache cache.Cache[RestoreStatus]
}

// NewRestoreCache creates a RestoreCache using the generic cache
func NewRestoreCache(cacheCfg config.CacheCfg) *RestoreCache {
	return &RestoreCache{
		cache: cache.CreateCache[RestoreStatus](cacheCfg),
	}
}

func (r *RestoreCache) getKey(key string) string {
	return fmt.Sprintf("mort:glacier:restore:%s", key)
}

// MarkRestoreRequested records that a restore was initiated
func (r *RestoreCache) MarkRestoreRequested(ctx context.Context, key string, expiresIn time.Duration) error {
	cacheKey := r.getKey(key)

	status := RestoreStatus{
		Key:         key,
		RequestedAt: time.Now(),
		ExpiresAt:   time.Now().Add(expiresIn),
		InProgress:  true,
	}

	// Set with TTL slightly longer than restore time to track completion
	return r.cache.Set(ctx, cacheKey, status, expiresIn+24*time.Hour)
}

// GetRestoreStatus checks if restore was already requested
func (r *RestoreCache) GetRestoreStatus(ctx context.Context, key string) (*RestoreStatus, error) {
	cacheKey := r.getKey(key)

	status, found, err := r.cache.Get(ctx, cacheKey)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}

	// Update InProgress based on current time
	status.InProgress = time.Now().Before(status.ExpiresAt)

	return &status, nil
}

// Global singleton restore cache
var restoreCache *RestoreCache
var restoreCacheOnce sync.Once

// GetRestoreCache returns the singleton restore cache instance
func GetRestoreCache(cacheCfg config.CacheCfg) *RestoreCache {
	restoreCacheOnce.Do(func() {
		restoreCache = NewRestoreCache(cacheCfg)
	})
	return restoreCache
}
