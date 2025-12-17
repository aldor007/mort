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

// testNotifyAndReleaseRedis is a test helper for redis tests (same as memory_test.go)
func testNotifyAndReleaseRedis(l Lock, ctx context.Context, key string, res *response.Response) {
	if res == nil {
		l.NotifyAndRelease(ctx, key, nil)
		return
	}

	sharedRes, err := response.NewSharedResponse(res)
	if err != nil {
		l.NotifyAndRelease(ctx, key, nil)
		return
	}
	l.NotifyAndRelease(ctx, key, sharedRes)
}

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

	go testNotifyAndReleaseRedis(l, ctx, key, response.NewError(400, errors.New("invalid transform")))

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
	go testNotifyAndReleaseRedis(l, ctx, key, response.NewBuf(200, buf))

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
	go testNotifyAndReleaseRedis(l, ctx, key, response.NewBuf(200, buf))

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
func TestRedisLock_Cancel(t *testing.T) {
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

	go func() {
		c.Cancel <- true
	}()

	buf := make([]byte, 1000)
	go testNotifyAndReleaseRedis(l, ctx, key, response.NewBuf(200, buf))

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

// TestRedisLock_PubSubCleanup verifies that pubsub connection is properly closed
// when a message is received from Redis pub/sub channel
func TestRedisLock_PubSubCleanup(t *testing.T) {
	t.Parallel()

	s := miniredis.RunT(t)

	// Create first lock instance that will acquire the lock
	l1 := NewRedisLock([]string{s.Addr()}, nil)
	key := "test-pubsub-cleanup"
	ctx := context.Background()

	// First instance acquires the lock
	_, acquired := l1.Lock(ctx, key)
	assert.True(t, acquired, "First instance should acquire lock")

	// Create second lock instance that will wait for the lock
	l2 := NewRedisLock([]string{s.Addr()}, nil)
	result, locked := l2.Lock(ctx, key)

	assert.False(t, locked, "Second instance shouldn't acquire lock")
	assert.NotNil(t, result.ResponseChan, "Should return response channel for waiting")

	// Verify that l2 has created a pubsub subscription
	l2.lock.RLock()
	lockData, exists := l2.locks[key]
	l2.lock.RUnlock()

	assert.True(t, exists, "Lock entry should exist in l2.locks map")
	assert.NotNil(t, lockData.pubsub, "PubSub should be initialized")

	// First instance notifies and releases (publishes to Redis)
	buf := make([]byte, 100)
	go testNotifyAndReleaseRedis(l1, ctx, key, response.NewBuf(200, buf))

	// Wait for the response to propagate via pubsub
	// In a two-instance scenario, the channel is closed (not sent a response)
	// because the actual response data doesn't transfer across instances
	timer := time.NewTimer(time.Second * 2)
	select {
	case <-timer.C:
		t.Fatalf("Timeout waiting for pubsub message")
		return
	case res, ok := <-result.ResponseChan:
		assert.False(t, ok, "Channel should be closed")
		assert.Nil(t, res, "Response should be nil for cross-instance notification")
	}

	// Give a small amount of time for cleanup to complete
	time.Sleep(50 * time.Millisecond)

	// Verify that pubsub connection and locks entry are cleaned up
	l2.lock.RLock()
	_, stillExists := l2.locks[key]
	l2.lock.RUnlock()

	assert.False(t, stillExists, "Lock entry should be deleted from l2.locks map after pubsub message")
}
