package cache

import (
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

func TestRedisCache_Set_minUse(t *testing.T) {
	db, mock := redismock.NewClientMock()

	i := RedisCache{
		cache: redisCache.New(&redisCache.Options{
			Redis: db,
		}),
		client: db,
		cfg:    CacheCfg{MinUseCount: 2},
	}

	obj := object.FileObject{}
	obj.Key = "cacheKey"
	res := response.NewString(200, "test")

	mock.ExpectIncr("countmort-v1:" + obj.GetResponseCacheKey()).SetVal(1)
	err := i.Set(&obj, res)
	assert.Nil(t, err)

	mock.ExpectIncr("countmort-v1:" + obj.GetResponseCacheKey()).SetVal(2)
	mock.ExpectSet("mort-v1:"+obj.GetResponseCacheKey(), res, time.Duration(res.GetTTL()))
	err = i.Set(&obj, res)
	assert.Nil(t, err)
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
