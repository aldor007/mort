package lock

import (
	"context"
	"errors"
	"runtime"
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

// TestRedisLock_AutoRefresh verifies that locks are automatically refreshed
// during long-running operations to prevent expiration
func TestRedisLock_AutoRefresh(t *testing.T) {
	t.Parallel()

	s := miniredis.RunT(t)

	// Create a lock with short timeout for faster testing
	l := NewRedisLock([]string{s.Addr()}, nil)
	l.LockTimeout = 2 // 2 seconds TTL, refresh every 1 second
	key := "test-auto-refresh"
	ctx := context.Background()

	// Acquire the lock
	_, acquired := l.Lock(ctx, key)
	assert.True(t, acquired, "Should acquire lock")

	// Verify refresh goroutine was started
	l.lock.RLock()
	lockData, exists := l.locks[key]
	l.lock.RUnlock()
	assert.True(t, exists, "Lock entry should exist")
	assert.NotNil(t, lockData.cancelRefresh, "Refresh cancel function should be set")

	// Simulate a long-running operation (3 seconds, longer than lock TTL)
	// The lock should be automatically refreshed and remain valid
	time.Sleep(3 * time.Second)

	// Verify the lock still exists in Redis (hasn't expired)
	// by checking if we can't acquire it with a second instance
	l2 := NewRedisLock([]string{s.Addr()}, nil)
	l2.LockTimeout = 2
	_, acquired2 := l2.Lock(ctx, key)
	assert.False(t, acquired2, "Second instance shouldn't acquire lock because it's still held and refreshed")

	// Release the lock
	l.Release(ctx, key)

	// Verify lock was released and refresh stopped
	l.lock.RLock()
	_, stillExists := l.locks[key]
	l.lock.RUnlock()
	assert.False(t, stillExists, "Lock entry should be deleted after release")

	// Create a fresh instance to acquire the lock
	// (l2 has a pubsub subscription from the previous attempt)
	l3 := NewRedisLock([]string{s.Addr()}, nil)
	l3.LockTimeout = 2
	_, acquired3 := l3.Lock(ctx, key)
	assert.True(t, acquired3, "Should acquire lock after release")

	l3.Release(ctx, key)
}

// TestRedisLock_RefreshStopsOnNotifyAndRelease verifies that refresh goroutine
// is stopped when NotifyAndRelease is called
func TestRedisLock_RefreshStopsOnNotifyAndRelease(t *testing.T) {
	t.Parallel()

	s := miniredis.RunT(t)

	l := NewRedisLock([]string{s.Addr()}, nil)
	l.LockTimeout = 2
	key := "test-refresh-notify"
	ctx := context.Background()

	// Acquire lock
	_, acquired := l.Lock(ctx, key)
	assert.True(t, acquired, "Should acquire lock")

	// Verify refresh is running
	l.lock.RLock()
	lockData, _ := l.locks[key]
	l.lock.RUnlock()
	assert.NotNil(t, lockData.cancelRefresh, "Refresh should be running")

	// Call NotifyAndRelease
	buf := make([]byte, 100)
	testNotifyAndReleaseRedis(l, ctx, key, response.NewBuf(200, buf))

	// Verify lock entry and refresh were cleaned up
	l.lock.RLock()
	_, exists := l.locks[key]
	l.lock.RUnlock()
	assert.False(t, exists, "Lock entry should be deleted after NotifyAndRelease")
}

// TestRedisLock_NoRefreshForWaitingClients verifies that waiting clients
// (those that didn't acquire the lock) don't start refresh goroutines
func TestRedisLock_NoRefreshForWaitingClients(t *testing.T) {
	t.Parallel()

	s := miniredis.RunT(t)

	l1 := NewRedisLock([]string{s.Addr()}, nil)
	l2 := NewRedisLock([]string{s.Addr()}, nil)
	key := "test-no-refresh-waiting"
	ctx := context.Background()

	// First instance acquires lock
	_, acquired := l1.Lock(ctx, key)
	assert.True(t, acquired, "First instance should acquire lock")

	// Verify first instance has refresh
	l1.lock.RLock()
	lockData1, _ := l1.locks[key]
	l1.lock.RUnlock()
	assert.NotNil(t, lockData1.cancelRefresh, "Lock holder should have refresh")

	// Second instance waits for lock
	_, acquired2 := l2.Lock(ctx, key)
	assert.False(t, acquired2, "Second instance should not acquire lock")

	// Verify second instance doesn't have refresh (has pubsub instead)
	l2.lock.RLock()
	lockData2, _ := l2.locks[key]
	l2.lock.RUnlock()
	assert.Nil(t, lockData2.cancelRefresh, "Waiting client should not have refresh")
	assert.NotNil(t, lockData2.pubsub, "Waiting client should have pubsub")

	// Cleanup
	buf := make([]byte, 100)
	testNotifyAndReleaseRedis(l1, ctx, key, response.NewBuf(200, buf))
}

// TestRedisLock_GoroutineLeak verifies that no goroutines are leaked
// after lock acquire, refresh, and release cycles
func TestRedisLock_GoroutineLeak(t *testing.T) {
	t.Parallel()

	s := miniredis.RunT(t)

	// Get baseline goroutine count
	runtime.GC()
	time.Sleep(200 * time.Millisecond)
	baselineGoroutines := runtime.NumGoroutine()

	l := NewRedisLock([]string{s.Addr()}, nil)
	l.LockTimeout = 2
	ctx := context.Background()

	// Perform multiple lock/release cycles
	for i := 0; i < 10; i++ {
		key := "test-leak-" + string(rune(i))

		// Acquire lock
		_, acquired := l.Lock(ctx, key)
		assert.True(t, acquired, "Should acquire lock")

		// Let refresh run at least once
		time.Sleep(1100 * time.Millisecond)

		// Release lock
		l.Release(ctx, key)
	}

	// Close the RedisLock to clean up Redis client goroutines
	err := l.Close()
	assert.NoError(t, err, "Close should not error")

	// Give more time for all goroutines to clean up
	time.Sleep(1 * time.Second)
	runtime.GC()
	time.Sleep(200 * time.Millisecond)

	// Check goroutine count
	finalGoroutines := runtime.NumGoroutine()

	// Allow some tolerance (5 goroutines) for background processes
	// Redis client might have some lingering goroutines that take time to shut down
	goroutineDiff := finalGoroutines - baselineGoroutines
	assert.LessOrEqual(t, goroutineDiff, 5,
		"Goroutine leak detected: baseline=%d, final=%d, diff=%d",
		baselineGoroutines, finalGoroutines, goroutineDiff)
}

// TestRedisLock_GoroutineLeakWithNotifyAndRelease verifies no goroutines leak
// when using NotifyAndRelease instead of Release
func TestRedisLock_GoroutineLeakWithNotifyAndRelease(t *testing.T) {
	t.Parallel()

	s := miniredis.RunT(t)

	// Get baseline goroutine count
	runtime.GC()
	time.Sleep(200 * time.Millisecond)
	baselineGoroutines := runtime.NumGoroutine()

	l := NewRedisLock([]string{s.Addr()}, nil)
	l.LockTimeout = 2
	ctx := context.Background()

	// Perform multiple lock/NotifyAndRelease cycles
	for i := 0; i < 10; i++ {
		key := "test-leak-notify-" + string(rune(i))

		// Acquire lock
		_, acquired := l.Lock(ctx, key)
		assert.True(t, acquired, "Should acquire lock")

		// Let refresh run at least once
		time.Sleep(1100 * time.Millisecond)

		// NotifyAndRelease
		buf := make([]byte, 100)
		testNotifyAndReleaseRedis(l, ctx, key, response.NewBuf(200, buf))
	}

	// Close the RedisLock to clean up Redis client goroutines
	err := l.Close()
	assert.NoError(t, err, "Close should not error")

	// Give more time for all goroutines to clean up
	time.Sleep(1 * time.Second)
	runtime.GC()
	time.Sleep(200 * time.Millisecond)

	// Check goroutine count
	finalGoroutines := runtime.NumGoroutine()

	// Allow some tolerance (5 goroutines) for background processes
	// Redis client might have some lingering goroutines that take time to shut down
	goroutineDiff := finalGoroutines - baselineGoroutines
	assert.LessOrEqual(t, goroutineDiff, 5,
		"Goroutine leak detected: baseline=%d, final=%d, diff=%d",
		baselineGoroutines, finalGoroutines, goroutineDiff)
}

// TestRedisLock_GoroutineLeakWithPubSub verifies no goroutines leak
// when using pubsub (waiting for locks held by other instances)
func TestRedisLock_GoroutineLeakWithPubSub(t *testing.T) {
	t.Parallel()

	s := miniredis.RunT(t)

	// Get baseline goroutine count
	runtime.GC()
	time.Sleep(200 * time.Millisecond)
	baselineGoroutines := runtime.NumGoroutine()

	ctx := context.Background()

	// Collect all lock instances for cleanup
	var locks []*RedisLock

	// Perform multiple cycles with two instances
	for i := 0; i < 5; i++ {
		key := "test-leak-pubsub-" + string(rune(i))

		l1 := NewRedisLock([]string{s.Addr()}, nil)
		l2 := NewRedisLock([]string{s.Addr()}, nil)
		locks = append(locks, l1, l2)

		// First instance acquires lock
		_, acquired := l1.Lock(ctx, key)
		assert.True(t, acquired, "First instance should acquire lock")

		// Second instance waits (creates pubsub)
		result, acquired2 := l2.Lock(ctx, key)
		assert.False(t, acquired2, "Second instance should wait")
		assert.NotNil(t, result.ResponseChan, "Should have response channel")

		// Release from first instance (triggers pubsub)
		buf := make([]byte, 100)
		go testNotifyAndReleaseRedis(l1, ctx, key, response.NewBuf(200, buf))

		// Wait for pubsub notification
		select {
		case <-result.ResponseChan:
			// Received notification
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout waiting for pubsub notification")
		}

		// Give time for cleanup
		time.Sleep(100 * time.Millisecond)
	}

	// Close all RedisLock instances to clean up Redis client goroutines
	for _, l := range locks {
		err := l.Close()
		assert.NoError(t, err, "Close should not error")
	}

	// Give more time for all goroutines to clean up
	// Multiple Redis clients need more cleanup time
	time.Sleep(1 * time.Second)
	runtime.GC()
	time.Sleep(200 * time.Millisecond)

	// Check goroutine count
	finalGoroutines := runtime.NumGoroutine()

	// Allow some tolerance (10 goroutines) for background processes
	// Multiple Redis clients might have lingering goroutines that take time to shut down
	goroutineDiff := finalGoroutines - baselineGoroutines
	assert.LessOrEqual(t, goroutineDiff, 10,
		"Goroutine leak detected: baseline=%d, final=%d, diff=%d",
		baselineGoroutines, finalGoroutines, goroutineDiff)
}
