package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

// testStruct is a msgpack-serializable type for testing
type testStruct struct {
	Name  string    `msgpack:"name"`
	Count int       `msgpack:"count"`
	When  time.Time `msgpack:"when"`
}

func TestGenericMemoryCache_SetGet(t *testing.T) {
	t.Parallel()

	cache := NewGenericMemoryCache[testStruct]()
	ctx := context.Background()

	value := testStruct{
		Name:  "test",
		Count: 42,
		When:  time.Now(),
	}

	// Set value
	err := cache.Set(ctx, "test-key", value, 1*time.Hour)
	assert.NoError(t, err, "Set should succeed")

	// Get value
	result, found, err := cache.Get(ctx, "test-key")
	assert.NoError(t, err, "Get should succeed")
	assert.True(t, found, "value should be found")
	assert.Equal(t, value.Name, result.Name, "Name should match")
	assert.Equal(t, value.Count, result.Count, "Count should match")
}

func TestGenericMemoryCache_Expiration(t *testing.T) {
	t.Parallel()

	cache := NewGenericMemoryCache[testStruct]()
	ctx := context.Background()

	value := testStruct{Name: "test", Count: 1}

	// Set with short TTL
	err := cache.Set(ctx, "expire-key", value, 50*time.Millisecond)
	assert.NoError(t, err)

	// Should be found immediately
	_, found, err := cache.Get(ctx, "expire-key")
	assert.NoError(t, err)
	assert.True(t, found, "should be found before expiration")

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should not be found after expiration
	_, found, err = cache.Get(ctx, "expire-key")
	assert.NoError(t, err)
	assert.False(t, found, "should not be found after expiration")
}

func TestGenericMemoryCache_Delete(t *testing.T) {
	t.Parallel()

	cache := NewGenericMemoryCache[testStruct]()
	ctx := context.Background()

	value := testStruct{Name: "test"}
	cache.Set(ctx, "delete-key", value, 0)

	// Verify it exists
	_, found, _ := cache.Get(ctx, "delete-key")
	assert.True(t, found)

	// Delete it
	err := cache.Delete(ctx, "delete-key")
	assert.NoError(t, err)

	// Should not be found
	_, found, _ = cache.Get(ctx, "delete-key")
	assert.False(t, found, "should not be found after delete")
}

func TestGenericRedisCache_SetGet(t *testing.T) {
	t.Parallel()

	// Create miniredis server
	s := miniredis.RunT(t)

	cache := NewGenericRedisCache[testStruct]([]string{s.Addr()}, nil, false)
	ctx := context.Background()

	value := testStruct{
		Name:  "redis-test",
		Count: 100,
		When:  time.Now().Truncate(time.Second), // Truncate for comparison
	}

	// Set value
	err := cache.Set(ctx, "redis-key", value, 1*time.Hour)
	assert.NoError(t, err, "Set should succeed")

	// Get value
	result, found, err := cache.Get(ctx, "redis-key")
	assert.NoError(t, err, "Get should succeed")
	assert.True(t, found, "value should be found")
	assert.Equal(t, value.Name, result.Name, "Name should match")
	assert.Equal(t, value.Count, result.Count, "Count should match")
	assert.Equal(t, value.When.Unix(), result.When.Unix(), "Time should match (Unix timestamp)")
}

func TestGenericRedisCache_Expiration(t *testing.T) {
	t.Parallel()

	s := miniredis.RunT(t)
	cache := NewGenericRedisCache[testStruct]([]string{s.Addr()}, nil, false)
	ctx := context.Background()

	value := testStruct{Name: "expire-test"}

	// Set with TTL
	err := cache.Set(ctx, "expire-key", value, 100*time.Millisecond)
	assert.NoError(t, err)

	// Should be found immediately
	_, found, err := cache.Get(ctx, "expire-key")
	assert.NoError(t, err)
	assert.True(t, found, "should be found before expiration")

	// Fast-forward time in miniredis
	s.FastForward(200 * time.Millisecond)

	// Should not be found after expiration
	_, found, err = cache.Get(ctx, "expire-key")
	assert.NoError(t, err)
	assert.False(t, found, "should not be found after expiration")
}

func TestGenericRedisCache_Delete(t *testing.T) {
	t.Parallel()

	s := miniredis.RunT(t)
	cache := NewGenericRedisCache[testStruct]([]string{s.Addr()}, nil, false)
	ctx := context.Background()

	value := testStruct{Name: "delete-test"}
	cache.Set(ctx, "delete-key", value, 0)

	// Delete it
	err := cache.Delete(ctx, "delete-key")
	assert.NoError(t, err)

	// Should not be found
	_, found, _ := cache.Get(ctx, "delete-key")
	assert.False(t, found, "should not be found after delete")
}

func TestGenericRedisCache_SharedConnectionPool(t *testing.T) {
	t.Parallel()

	s := miniredis.RunT(t)

	// Create two caches with same config
	cache1 := NewGenericRedisCache[testStruct]([]string{s.Addr()}, nil, false)
	cache2 := NewGenericRedisCache[testStruct]([]string{s.Addr()}, nil, false)

	// Both should use the same underlying Redis client
	redisCache1 := cache1.(*GenericRedisCache[testStruct])
	redisCache2 := cache2.(*GenericRedisCache[testStruct])

	assert.Same(t, redisCache1.client, redisCache2.client,
		"caches with same config should share Redis client")
}

func TestGenericRedisCache_DifferentTypes(t *testing.T) {
	t.Parallel()

	s := miniredis.RunT(t)
	ctx := context.Background()

	// Create cache for testStruct
	cache1 := NewGenericRedisCache[testStruct]([]string{s.Addr()}, nil, false)

	// Create cache for string
	cache2 := NewGenericRedisCache[string]([]string{s.Addr()}, nil, false)

	// Both should share the same client (connection pool)
	redisCache1 := cache1.(*GenericRedisCache[testStruct])
	redisCache2 := cache2.(*GenericRedisCache[string])

	assert.Same(t, redisCache1.client, redisCache2.client,
		"caches with different types but same config should share client")

	// Store different types
	cache1.Set(ctx, "struct-key", testStruct{Name: "test"}, 1*time.Hour)
	cache2.Set(ctx, "string-key", "test-string", 1*time.Hour)

	// Both should work independently
	val1, found1, _ := cache1.Get(ctx, "struct-key")
	assert.True(t, found1)
	assert.Equal(t, "test", val1.Name)

	val2, found2, _ := cache2.Get(ctx, "string-key")
	assert.True(t, found2)
	assert.Equal(t, "test-string", val2)
}

func TestGenericCache_MsgpackSerialization(t *testing.T) {
	t.Parallel()

	s := miniredis.RunT(t)
	cache := NewGenericRedisCache[testStruct]([]string{s.Addr()}, nil, false)
	ctx := context.Background()

	// Complex struct with all field types
	value := testStruct{
		Name:  "msgpack-test",
		Count: 12345,
		When:  time.Date(2025, 12, 19, 10, 30, 0, 0, time.UTC),
	}

	// Set and get
	err := cache.Set(ctx, "msgpack-key", value, 1*time.Hour)
	assert.NoError(t, err, "msgpack serialization should succeed")

	result, found, err := cache.Get(ctx, "msgpack-key")
	assert.NoError(t, err, "msgpack deserialization should succeed")
	assert.True(t, found)
	assert.Equal(t, value.Name, result.Name)
	assert.Equal(t, value.Count, result.Count)
	assert.Equal(t, value.When.Unix(), result.When.Unix())
}
