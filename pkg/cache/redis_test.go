package cache

import (
	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/response"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type mockValue struct {
	value      interface{}
	expiration time.Duration
}

var internal map[string]mockValue

type redisMock struct {
}

func (m redisMock) Set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	internal[key] = mockValue{value, expiration}
	return redis.NewStatusCmd("ok")
}
func (m redisMock) Get(key string) *redis.StringCmd {
	d, ok := internal[key]
	if !ok {
		return redis.NewStringCmd("nil")
	}

	r := redis.NewStringCmd(d)
	return r
}

func (m redisMock) Del(keys ...string) *redis.IntCmd {
	delete(internal, keys[0])
	return redis.NewIntResult(0, nil)
}

func TestRedisCache_Get_and_Set(t *testing.T) {
	r := NewRedis([]string{"localhost:1"}, nil)

	m := redisMock{}
	internal = make(map[string]mockValue)
	r.client.Redis = m

	obj := &object.FileObject{}
	obj.Key = "key"

	res := response.NewString(300, "test")
	res.Headers.Set("cache-control", "public, max-age=60")

	r.Set(obj, res)
	res2, err := r.Get(obj)

	assert.Nil(t, err)
	assert.Equal(t, res2.StatusCode, 300)
	assert.Equal(t, res2.GetTTL(), 60)

	internalItem := internal[obj.GetResponseCacheKey()]
	assert.Equal(t, internalItem.expiration, time.Second*time.Duration(res.GetTTL()))

	r.Delete(obj)
	_, err = r.Get(obj)

}
