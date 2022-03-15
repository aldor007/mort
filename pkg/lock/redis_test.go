package lock

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aldor007/mort/pkg/response"
	"github.com/bsm/redislock"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
)

func TestNewRedisLock(t *testing.T) {
	l := NewRedisLock([]string{"1.1.1.1:1234"}, nil)

	assert.NotNil(t, l, "New Redis should return not nil")
}

func TestRedisLock_Lock(t *testing.T) {
	l := NewRedisLock([]string{"1.1.1.1:1234"}, nil)
	client, mock := redismock.NewClientMock()
	l.client = redislock.New(client)
	key := "klucz"
	mock.Regexp().ExpectSetNX(key, `[a-zA-Z0-9]+`, 60*time.Second).SetVal(true)
	ctx := context.Background()
	c, acquired := l.Lock(ctx, key)

	assert.True(t, acquired, "Should acquire lock")
	assert.Nil(t, c.ResponseChan, "shouldn't return channel")
	assert.Nil(t, c.Error, "error should be nil")

	mock.ClearExpect()
	mock.Regexp().ExpectSetNX(key, `[a-zA-Z0-9]+`, 60*time.Second).SetVal(false)
	resChan, lock := l.Lock(ctx, key)

	assert.False(t, lock, "Shouldn't acquire lock")
	assert.NotNil(t, resChan, "should return channel")

	mock.ClearExpect()
	mock.Regexp().ExpectSetNX(key, `[a-zA-Z0-9]+`, 60*time.Second).SetVal(true)
	l.Release(ctx, key)

	c, acquired = l.Lock(ctx, key)

	assert.True(t, acquired, "Should acquire lock after release")
	assert.Nil(t, c.ResponseChan, "shouldn't return channel after release")
	assert.Nil(t, c.Error, "error should be nil")
}

func TestRedisLock_NotifyAndReleaseWhenError(t *testing.T) {
	l := NewRedisLock([]string{"1.1.1.1:1234"}, nil)
	client, mock := redismock.NewClientMock()
	l.client = redislock.New(client)
	key := "kluczi2"
	ctx := context.Background()
	mock.Regexp().ExpectSetNX(key, `[a-zA-Z0-9]+`, 60*time.Second).SetVal(true)
	c, acquired := l.Lock(ctx, key)

	assert.True(t, acquired, "Should acquire lock")
	assert.Nil(t, c.ResponseChan, "shouldn't return channel")
	assert.Nil(t, c.Error, "error should be nil")

	mock.ClearExpect()
	mock.Regexp().ExpectSetNX(key, `[a-zA-Z0-9]+`, 60*time.Second).SetVal(false)
	result, lock := l.Lock(ctx, key)

	assert.False(t, lock, "Shouldn't acquire lock")
	assert.NotNil(t, result, "should return channel")

	go l.NotifyAndRelease(ctx, key, response.NewError(400, errors.New("invalid transform")))

	timer := time.NewTimer(time.Second * 2)
	select {
	case <-timer.C:
		t.Fatalf("Timeout while waiting for response propagation")
		return
	case res := <-result.ResponseChan:
		assert.NotNil(t, res, "Response shouldn't be nil")
		if res != nil {
			assert.Equal(t, res.StatusCode, 400)
		}
	}
}

func TestRedisLock_NotifyAndRelease(t *testing.T) {
	l := NewRedisLock([]string{"1.1.1.1:1234"}, nil)
	client, mock := redismock.NewClientMock()
	l.client = redislock.New(client)
	key := "kluczi22"
	ctx := context.Background()
	mock.ClearExpect()
	mock.Regexp().ExpectSetNX(key, `[a-zA-Z0-9]+`, 60*time.Second).SetVal(true)
	c, acquired := l.Lock(ctx, key)

	assert.True(t, acquired, "Should acquire lock")
	assert.Nil(t, c.ResponseChan, "shouldn't return channel")

	mock.ClearExpect()
	mock.Regexp().ExpectSetNX(key, `[a-zA-Z0-9]+`, 60*time.Second).SetVal(false)
	result, lock := l.Lock(ctx, key)

	assert.False(t, lock, "Shouldn't acquire lock")
	assert.NotNil(t, result.ResponseChan, "should return channel")

	buf := make([]byte, 1000)
	go l.NotifyAndRelease(ctx, key, response.NewBuf(200, buf))

	timer := time.NewTimer(time.Second * 2)
	select {
	case <-timer.C:
		t.Fatalf("timeout waiting for lock")
		return
	case res := <-result.ResponseChan:
		assert.NotNil(t, res, "Response should't be nil")
		if res != nil {
			assert.Equal(t, res.StatusCode, 200, "Response should have sc = 200")
		}
		buf2, err := res.Body()
		assert.Nil(t, err)
		assert.Equal(t, len(buf), len(buf2))

	}
}

func TestRedisLock_NotifyAndRelease2(t *testing.T) {
	l := NewRedisCluster([]string{"1.1.1.1:1234"}, nil)
	client, mock := redismock.NewClientMock()
	l.client = redislock.New(client)
	key := "kluczi222"
	ctx := context.Background()
	mock.ClearExpect()
	mock.Regexp().ExpectSetNX(key, `[a-zA-Z0-9]+`, 60*time.Second).SetVal(true)
	c, acquired := l.Lock(ctx, key)

	assert.True(t, acquired, "Should acquire lock")
	assert.Nil(t, c.ResponseChan, "shouldn't return channel")

	l.NotifyAndRelease(ctx, "no-key", response.NewError(400, errors.New("invalid transform")))
	l.NotifyAndRelease(ctx, key, response.NewNoContent(200))
}
