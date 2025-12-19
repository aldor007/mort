package cache

import (
	"testing"

	"github.com/aldor007/mort/pkg/config"
	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

// TestCreateCache_MultipleCallsShareRedisClient verifies connection pool sharing
func TestCreateCache_MultipleCallsShareRedisClient(t *testing.T) {
	t.Parallel()

	s := miniredis.RunT(t)

	cfg := config.CacheCfg{
		Type:    "redis",
		Address: []string{s.Addr()},
	}

	// Create first cache
	cache1 := CreateCache[testStruct](cfg)
	redisCache1, ok1 := cache1.(*GenericRedisCache[testStruct])
	assert.True(t, ok1, "should be GenericRedisCache")

	// Create second cache with same config
	cache2 := CreateCache[testStruct](cfg)
	redisCache2, ok2 := cache2.(*GenericRedisCache[testStruct])
	assert.True(t, ok2, "should be GenericRedisCache")

	// Different Cache instances
	assert.NotSame(t, cache1, cache2, "Cache instances are different (OK)")

	// But same underlying Redis client (shared pool)
	assert.Same(t, redisCache1.client, redisCache2.client,
		"Redis clients should be the same (shared pool)")
}

// TestCreateCache_DifferentTypesShareClient verifies pool sharing across types
func TestCreateCache_DifferentTypesShareClient(t *testing.T) {
	t.Parallel()

	s := miniredis.RunT(t)

	cfg := config.CacheCfg{
		Type:    "redis",
		Address: []string{s.Addr()},
	}

	// Create cache for different types
	cache1 := CreateCache[testStruct](cfg)
	cache2 := CreateCache[string](cfg)

	redisCache1 := cache1.(*GenericRedisCache[testStruct])
	redisCache2 := cache2.(*GenericRedisCache[string])

	// Should share the same Redis client despite different types
	assert.Same(t, redisCache1.client, redisCache2.client,
		"Different Cache[T] types should share Redis client")
}

// TestCreateCache_MemoryVsRedis verifies backend selection
func TestCreateCache_MemoryVsRedis(t *testing.T) {
	t.Parallel()

	s := miniredis.RunT(t)

	memCfg := config.CacheCfg{Type: "memory"}
	redisCfg := config.CacheCfg{Type: "redis", Address: []string{s.Addr()}}

	memCache := CreateCache[testStruct](memCfg)
	redisCache := CreateCache[testStruct](redisCfg)

	_, isMemory := memCache.(*GenericMemoryCache[testStruct])
	_, isRedis := redisCache.(*GenericRedisCache[testStruct])

	assert.True(t, isMemory, "should create memory cache for type=memory")
	assert.True(t, isRedis, "should create redis cache for type=redis")
}
