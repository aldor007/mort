package cache

import (
	"testing"

	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/response"
	"github.com/stretchr/testify/assert"
	"github.com/go-redis/redismock/v8"
	redisCache "github.com/go-redis/cache/v8"
	"time"
)

func TestRedisCache_Set(t *testing.T) {

	db, mock := redismock.NewClientMock()

	i := RedisCache{
		client:  redisCache.New(&redisCache.Options{
		Redis:      db,
		}),
	}

	obj := object.FileObject{}
	obj.Key = "cacheKey"
	res := response.NewString(200, "test")

	mock.ExpectSet("mort-v1:" + obj.GetResponseCacheKey(), res, time.Duration(res.GetTTL()))
	err := i.Set(&obj, res)
	assert.Nil(t, err)

}

func TestRedisCache_Delete(t *testing.T) {
	db, mock := redismock.NewClientMock()

	i := RedisCache{
		client:  redisCache.New(&redisCache.Options{
		Redis:      db,
		}),
	}

	obj := object.FileObject{}
	obj.Key = "cacheKey"
	res := response.NewString(200, "test")

	mock.ExpectDel("mort-v1:" + obj.GetResponseCacheKey())
	err := i.Set(&obj, res)
	assert.Nil(t, err)
}

func TestRedisCache_Get(t *testing.T) {

	db, mock := redismock.NewClientMock()

	i := RedisCache{
		client:  redisCache.New(&redisCache.Options{
		Redis:      db,
		}),
	}

	obj := object.FileObject{}
	obj.Key = "cacheKey"

	mock.ExpectGet("mort-v1:" + obj.GetResponseCacheKey()).RedisNil()
	_, err := i.Get(&obj)
	assert.NotNil(t, err)

}