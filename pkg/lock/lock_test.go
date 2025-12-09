package lock

import (
	"context"
	"testing"

	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/response"
	"github.com/stretchr/testify/assert"
)

func TestNewNopLock(t *testing.T) {
	l := NewNopLock()

	ctx := context.Background()
	_, locked := l.Lock(ctx, "a")
	assert.True(t, locked)

	_, locked = l.Lock(ctx, "a")
	assert.True(t, locked)

	l.NotifyAndRelease(ctx, "a", response.NewNoContent(200))

	_, locked = l.Lock(ctx, "a")
	assert.True(t, locked)

	l.Release(ctx, "a")

	_, locked = l.Lock(ctx, "a")
	assert.True(t, locked)
}

func TestCreate_NilConfig(t *testing.T) {
	t.Parallel()

	l := Create(nil, 30)
	assert.NotNil(t, l)

	// Should return a memory lock
	_, ok := l.(*MemoryLock)
	assert.True(t, ok, "Should create MemoryLock when config is nil")
}

func TestCreate_DefaultType(t *testing.T) {
	t.Parallel()

	cfg := &config.LockCfg{
		Type: "unknown",
	}
	l := Create(cfg, 30)
	assert.NotNil(t, l)

	// Should return a memory lock for unknown type
	_, ok := l.(*MemoryLock)
	assert.True(t, ok, "Should create MemoryLock for unknown type")
}

func TestCreate_RedisType(t *testing.T) {
	t.Parallel()

	cfg := &config.LockCfg{
		Type:    "redis",
		Address: []string{"localhost:6379"},
	}
	l := Create(cfg, 45)
	assert.NotNil(t, l)

	// Should return a redis lock
	rl, ok := l.(*RedisLock)
	assert.True(t, ok, "Should create RedisLock for redis type")
	assert.Equal(t, 45, rl.LockTimeout, "Should set lock timeout")
}

func TestCreate_RedisClusterType(t *testing.T) {
	t.Parallel()

	cfg := &config.LockCfg{
		Type:    "redis-cluster",
		Address: []string{"localhost:7000", "localhost:7001"},
	}
	l := Create(cfg, 60)
	assert.NotNil(t, l)

	// Should return a redis lock
	rl, ok := l.(*RedisLock)
	assert.True(t, ok, "Should create RedisLock for redis-cluster type")
	assert.Equal(t, 60, rl.LockTimeout, "Should set lock timeout")
}
