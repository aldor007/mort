package storage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aldor007/mort/pkg/config"
	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func TestRestoreCache_MarkAndGetStatus(t *testing.T) {
	t.Parallel()

	cfg := config.CacheCfg{Type: "memory"}
	cache := NewRestoreCache(cfg)
	ctx := context.Background()

	// Mark restore as requested
	err := cache.MarkRestoreRequested(ctx, "test/image.jpg", 1*time.Hour)
	assert.NoError(t, err, "should mark restore as requested")

	// Get status
	status, err := cache.GetRestoreStatus(ctx, "test/image.jpg")
	assert.NoError(t, err, "should get status")
	assert.NotNil(t, status, "status should exist")
	assert.Equal(t, "test/image.jpg", status.Key)
	assert.True(t, status.InProgress, "should be in progress")
	assert.WithinDuration(t, time.Now(), status.RequestedAt, 1*time.Second)
}

func TestRestoreCache_Expiration(t *testing.T) {
	t.Parallel()

	cfg := config.CacheCfg{Type: "memory"}
	cache := NewRestoreCache(cfg)
	ctx := context.Background()

	// Mark with short TTL
	err := cache.MarkRestoreRequested(ctx, "expire/image.jpg", 50*time.Millisecond)
	assert.NoError(t, err)

	// Should be found immediately
	status, err := cache.GetRestoreStatus(ctx, "expire/image.jpg")
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.True(t, status.InProgress)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should not be found (or not in progress) after expiration
	status, err = cache.GetRestoreStatus(ctx, "expire/image.jpg")
	assert.NoError(t, err)
	if status != nil {
		assert.False(t, status.InProgress, "should not be in progress after expiration")
	}
}

func TestRestoreCache_NotFound(t *testing.T) {
	t.Parallel()

	cfg := config.CacheCfg{Type: "memory"}
	cache := NewRestoreCache(cfg)
	ctx := context.Background()

	// Get non-existent key
	status, err := cache.GetRestoreStatus(ctx, "nonexistent/image.jpg")
	assert.NoError(t, err, "should not error on missing key")
	assert.Nil(t, status, "status should be nil for non-existent key")
}

func TestRestoreCache_WithRedis(t *testing.T) {
	t.Parallel()

	s := miniredis.RunT(t)

	cfg := config.CacheCfg{
		Type:    "redis",
		Address: []string{s.Addr()},
	}
	cache := NewRestoreCache(cfg)
	ctx := context.Background()

	// Mark restore
	err := cache.MarkRestoreRequested(ctx, "redis/image.jpg", 1*time.Hour)
	assert.NoError(t, err)

	// Get status
	status, err := cache.GetRestoreStatus(ctx, "redis/image.jpg")
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, "redis/image.jpg", status.Key)
	assert.True(t, status.InProgress)
}

func TestRestoreCache_KeyPrefix(t *testing.T) {
	t.Parallel()

	cfg := config.CacheCfg{Type: "memory"}
	cache := NewRestoreCache(cfg)

	// Verify key prefix
	key := cache.getKey("test.jpg")
	assert.Equal(t, "mort:glacier:restore:test.jpg", key,
		"should use mort:glacier:restore: prefix")
}

func TestRestoreCache_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	cfg := config.CacheCfg{Type: "memory"}
	cache := NewRestoreCache(cfg)
	ctx := context.Background()

	// Concurrent writes
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			key := fmt.Sprintf("concurrent/image-%d.jpg", id)
			cache.MarkRestoreRequested(ctx, key, 1*time.Hour)
			done <- true
		}(i)
	}

	// Wait for all writes
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all were stored
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("concurrent/image-%d.jpg", i)
		status, err := cache.GetRestoreStatus(ctx, key)
		assert.NoError(t, err)
		assert.NotNil(t, status, "status should exist for key %s", key)
	}
}

func TestGetRestoreCache_Singleton(t *testing.T) {
	// Note: Not using t.Parallel() because this tests global singleton

	cfg := config.CacheCfg{Type: "memory"}

	// Get cache twice
	cache1 := GetRestoreCache(cfg)
	cache2 := GetRestoreCache(cfg)

	// Should be the same instance (singleton)
	assert.Same(t, cache1, cache2, "GetRestoreCache should return singleton")
}

func TestRestoreCache_UpdateInProgressStatus(t *testing.T) {
	t.Parallel()

	cfg := config.CacheCfg{Type: "memory"}
	cache := NewRestoreCache(cfg)
	ctx := context.Background()

	// Mark with 100ms TTL
	cache.MarkRestoreRequested(ctx, "test.jpg", 100*time.Millisecond)

	// Should be in progress initially
	status1, _ := cache.GetRestoreStatus(ctx, "test.jpg")
	assert.NotNil(t, status1)
	assert.True(t, status1.InProgress)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should not be in progress after expiration
	status2, _ := cache.GetRestoreStatus(ctx, "test.jpg")
	if status2 != nil {
		assert.False(t, status2.InProgress,
			"InProgress should be false after ExpiresAt")
	}
}
