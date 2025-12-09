package cache

import (
	"context"
	"testing"

	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/response"
	redisCache "github.com/go-redis/cache/v8"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
	"time"
)

func TestRedisCache_Set(t *testing.T) {

	db, mock := redismock.NewClientMock()

	i := RedisCache{
		cache: redisCache.New(&redisCache.Options{
			Redis: db,
		}),
		client: db,
	}

	obj := object.FileObject{}
	obj.Key = "cacheKey"
	res := response.NewString(200, "test")

	mock.ExpectSet("mort-v1:"+obj.GetResponseCacheKey(), res, time.Duration(res.GetTTL()))
	err := i.Set(&obj, res)
	assert.Nil(t, err)

}

func TestRedisCache_Set_minUseCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		minUseCount uint64
		setCalls    int
		shouldIncr  bool
		shouldDel   bool
		description string
	}{
		{
			name:        "should not cache when minUseCount is 2 and called once",
			minUseCount: 2,
			setCalls:    1,
			shouldIncr:  true,
			shouldDel:   false,
			description: "first call should increment counter but not cache",
		},
		{
			name:        "should cache when minUseCount is 2 and called twice",
			minUseCount: 2,
			setCalls:    2,
			shouldIncr:  true,
			shouldDel:   true,
			description: "second call should reach threshold and cache",
		},
		{
			name:        "should cache when minUseCount is 1 and called once",
			minUseCount: 1,
			setCalls:    1,
			shouldIncr:  true,
			shouldDel:   true,
			description: "first call should reach threshold and cache immediately",
		},
		{
			name:        "should not cache when minUseCount is 3 and called twice",
			minUseCount: 3,
			setCalls:    2,
			shouldIncr:  true,
			shouldDel:   false,
			description: "two calls should not reach threshold of 3",
		},
		{
			name:        "should cache when minUseCount is 3 and called three times",
			minUseCount: 3,
			setCalls:    3,
			shouldIncr:  true,
			shouldDel:   true,
			description: "third call should reach threshold and cache",
		},
		{
			name:        "should cache when minUseCount is 5 and called five times",
			minUseCount: 5,
			setCalls:    5,
			shouldIncr:  true,
			shouldDel:   true,
			description: "fifth call should reach threshold and cache",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, mock := redismock.NewClientMock()

			cache := RedisCache{
				cache: redisCache.New(&redisCache.Options{
					Redis: db,
				}),
				client: db,
				cfg: CacheCfg{
					MinUseCount: tt.minUseCount,
					MaxItemSize: 10000, // Set a reasonable max size
				},
			}

			obj := object.FileObject{}
			obj.Key = "cacheKey"
			obj.Ctx = context.Background()
			res := response.NewString(200, "test")

			countKey := "countmort-v1:" + obj.GetResponseCacheKey()

			// Set up expectations for each call
			for i := uint64(1); i <= uint64(tt.setCalls); i++ {
				mock.ExpectIncr(countKey).SetVal(int64(i))

				// If this call reaches the threshold, expect counter deletion
				if i == tt.minUseCount {
					mock.ExpectDel(countKey).SetVal(1)
				}
			}

			// Execute Set calls
			for i := 0; i < tt.setCalls; i++ {
				_ = cache.Set(&obj, res)
			}

			// Verify all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestRedisCache_Set_minUseCountWithMaxItemSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		minUseCount uint64
		maxItemSize int64
		content     string
		setCalls    int
		expectIncr  bool
		description string
	}{
		{
			name:        "should not increment counter for large items",
			minUseCount: 2,
			maxItemSize: 5,
			content:     "this is too large",
			setCalls:    2,
			expectIncr:  false,
			description: "large items should be skipped before checking minUseCount",
		},
		{
			name:        "should increment counter and cache small items",
			minUseCount: 2,
			maxItemSize: 1000,
			content:     "small",
			setCalls:    2,
			expectIncr:  true,
			description: "small items should be cached after reaching threshold",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, mock := redismock.NewClientMock()

			cache := RedisCache{
				cache: redisCache.New(&redisCache.Options{
					Redis: db,
				}),
				client: db,
				cfg: CacheCfg{
					MinUseCount: tt.minUseCount,
					MaxItemSize: tt.maxItemSize,
				},
			}

			obj := object.FileObject{}
			obj.Key = "cacheKey"
			obj.Ctx = context.Background()
			res := response.NewString(200, tt.content)

			countKey := "countmort-v1:" + obj.GetResponseCacheKey()

			// Only set up expectations if item is small enough to consider caching
			if tt.expectIncr {
				for i := uint64(1); i <= uint64(tt.setCalls); i++ {
					mock.ExpectIncr(countKey).SetVal(int64(i))

					if i == tt.minUseCount {
						mock.ExpectDel(countKey).SetVal(1)
					}
				}
			}

			// Execute Set calls
			for i := 0; i < tt.setCalls; i++ {
				_ = cache.Set(&obj, res)
			}

			// Verify all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestRedisCache_Set_minUseCountZero(t *testing.T) {
	t.Parallel()

	db, _ := redismock.NewClientMock()

	i := RedisCache{
		cache: redisCache.New(&redisCache.Options{
			Redis: db,
		}),
		client: db,
		cfg:    CacheCfg{MinUseCount: 0},
	}

	obj := object.FileObject{}
	obj.Key = "cacheKey"
	obj.Ctx = context.Background()
	res := response.NewString(200, "test")

	// When MinUseCount is 0, should cache immediately without counter
	// No ExpectIncr or ExpectDel should be called
	err := i.Set(&obj, res)
	assert.Nil(t, err, "should cache immediately when minUseCount is 0")
}

func TestRedisCache_Delete(t *testing.T) {
	db, mock := redismock.NewClientMock()

	i := RedisCache{
		cache: redisCache.New(&redisCache.Options{
			Redis: db,
		}),
		client: db,
	}

	obj := object.FileObject{}
	obj.Key = "cacheKey"
	res := response.NewString(200, "test")

	mock.ExpectDel("mort-v1:" + obj.GetResponseCacheKey())
	err := i.Set(&obj, res)
	assert.Nil(t, err)
}

func TestRedisCache_Get(t *testing.T) {
	t.Parallel()

	db, mock := redismock.NewClientMock()

	i := RedisCache{
		cache: redisCache.New(&redisCache.Options{
			Redis: db,
		}),
		client: db,
	}

	obj := object.FileObject{}
	obj.Key = "cacheKey"

	mock.ExpectGet("mort-v1:" + obj.GetResponseCacheKey()).RedisNil()
	_, err := i.Get(&obj)
	assert.NotNil(t, err)
}

func TestRedisCache_SetTooLarge(t *testing.T) {
	t.Parallel()

	db, _ := redismock.NewClientMock()

	i := RedisCache{
		cache: redisCache.New(&redisCache.Options{
			Redis: db,
		}),
		client: db,
		cfg:    CacheCfg{MaxItemSize: 10}, // 10 bytes max
	}

	obj := object.FileObject{}
	obj.Key = "cacheKey"
	res := response.NewString(200, "this is a very long response that exceeds the limit")

	err := i.Set(&obj, res)
	assert.Nil(t, err) // Should return nil but not cache
}

func TestRedisCache_getKey(t *testing.T) {
	t.Parallel()

	i := RedisCache{}
	obj := object.FileObject{
		Bucket: "test-bucket",
		Key:    "test-key.jpg",
		Range:  "",
	}

	key := i.getKey(&obj)
	assert.Contains(t, key, "mort-v1:")
	assert.Contains(t, key, "test-bucket")
	assert.Contains(t, key, "test-key.jpg")
}

func TestParseAddress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		addrs    []string
		expected map[string]string
	}{
		{
			name:  "single address",
			addrs: []string{"localhost:6379"},
			expected: map[string]string{
				"localhost": "localhost:6379",
			},
		},
		{
			name:  "multiple addresses",
			addrs: []string{"host1:6379", "host2:6379"},
			expected: map[string]string{
				"host1": "host1:6379",
				"host2": "host2:6379",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAddress(tt.addrs)
			assert.Equal(t, tt.expected, result)
		})
	}
}
