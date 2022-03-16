package lock

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aldor007/mort/pkg/response"
	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewRedisLock(t *testing.T) {
	l := NewRedisLock([]string{"1.1.1.1:1234"}, nil)

	assert.NotNil(t, l, "New Redis should return not nil")
}

func TestRedisLock_Lock(t *testing.T) {
	s := miniredis.RunT(t)

	l := NewRedisCluster([]string{s.Addr()}, nil)
	key := "klucz"
	ctx := context.Background()
	c, acquired := l.Lock(ctx, key)

	assert.True(t, acquired, "Should acquire lock")
	assert.Nil(t, c.ResponseChan, "shouldn't return channel")
	assert.Nil(t, c.Error, "error should be nil")

	resChan, lock := l.Lock(ctx, key)

	assert.False(t, lock, "Shouldn't acquire lock")
	assert.NotNil(t, resChan, "should return channel")

	l.Release(ctx, key)

	c, acquired = l.Lock(ctx, key)

	assert.True(t, acquired, "Should acquire lock after release")
	assert.Nil(t, c.ResponseChan, "shouldn't return channel after release")
	assert.Nil(t, c.Error, "error should be nil")
}

func TestRedisLock_NotifyAndReleaseWhenError(t *testing.T) {
	s := miniredis.RunT(t)

	l := NewRedisLock([]string{s.Addr()}, nil)
	key := "kluczi2"
	ctx := context.Background()
	c, acquired := l.Lock(ctx, key)

	assert.True(t, acquired, "Should acquire lock")
	assert.Nil(t, c.ResponseChan, "shouldn't return channel")
	assert.Nil(t, c.Error, "error should be nil")

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
	s := miniredis.RunT(t)

	l := NewRedisLock([]string{s.Addr()}, nil)
	key := "kluczi22"
	ctx := context.Background()
	c, acquired := l.Lock(ctx, key)

	assert.True(t, acquired, "Should acquire lock")
	assert.Nil(t, c.ResponseChan, "shouldn't return channel")

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

func TestRedisLock_NotifyAndReleaseTwoInstancesOfLock(t *testing.T) {
	s := miniredis.RunT(t)

	l := NewRedisLock([]string{s.Addr()}, nil)
	key := "kluczi22"
	ctx := context.Background()
	c, acquired := l.Lock(ctx, key)

	assert.True(t, acquired, "Should acquire lock")
	assert.Nil(t, c.ResponseChan, "shouldn't return channel")

	l2 := NewRedisLock([]string{s.Addr()}, nil)
	result, lock := l2.Lock(ctx, key)

	assert.False(t, lock, "Shouldn't acquire lock")
	assert.NotNil(t, result.ResponseChan, "should return channel")

	buf := make([]byte, 1000)
	go l.NotifyAndRelease(ctx, key, response.NewBuf(200, buf))

	timer := time.NewTimer(time.Second * 2)
	select {
	case <-timer.C:
		t.Fatalf("timeout waiting for lock")
		return
	case res, ok := <-result.ResponseChan:
		assert.False(t, ok, "channel shouldn't be closed")
		assert.Nil(t, res, "should be nil")

	}
}
