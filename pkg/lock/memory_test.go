package lock

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aldor007/mort/pkg/response"
	"github.com/stretchr/testify/assert"
)

func TestNewMemoryLock(t *testing.T) {
	l := NewMemoryLock()

	assert.NotNil(t, l, "New memory should return not nil")
}

func TestMemoryLock_Lock(t *testing.T) {
	l := NewMemoryLock()
	key := "klucz"
	ctx := context.Background()
	c, acquired := l.Lock(ctx, key)

	assert.True(t, acquired, "Should acquire lock")
	assert.Nil(t, c.ResponseChan, "shouldn't return channel")

	resChan, lock := l.Lock(ctx, key)

	assert.False(t, lock, "Shouldn't acquire lock")
	assert.NotNil(t, resChan, "should return channel")

	l.Release(ctx, key)

	c, acquired = l.Lock(ctx, key)

	assert.True(t, acquired, "Should acquire lock after release")
	assert.Nil(t, c.ResponseChan, "shouldn't return channel after release")
}

func TestMemoryLock_NotifyAndReleaseWhenError(t *testing.T) {
	l := NewMemoryLock()
	key := "kluczi2"
	ctx := context.Background()
	c, acquired := l.Lock(ctx, key)

	assert.True(t, acquired, "Should acquire lock")
	assert.Nil(t, c.ResponseChan, "shouldn't return channel")

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

func TestMemoryLock_NotifyAndRelease(t *testing.T) {
	l := NewMemoryLock()
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
		t.Fatalf("timeout waitgin for lock")
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

func TestMemoryLock_NotifyAndRelease2(t *testing.T) {
	l := NewMemoryLock()
	key := "kluczi222"
	ctx := context.Background()
	c, acquired := l.Lock(ctx, key)

	assert.True(t, acquired, "Should acquire lock")
	assert.Nil(t, c.ResponseChan, "shouldn't return channel")

	l.NotifyAndRelease(ctx, "no-key", response.NewError(400, errors.New("invalid transform")))
	l.NotifyAndRelease(ctx, key, response.NewNoContent(200))
}

func TestMemoryLock_Release(t *testing.T) {
	t.Parallel()

	l := NewMemoryLock()
	key := "test-key"
	ctx := context.Background()

	// Acquire lock
	_, acquired := l.Lock(ctx, key)
	assert.True(t, acquired, "Should acquire lock")

	// Release lock
	l.Release(ctx, key)

	// Should be able to acquire again
	_, acquired = l.Lock(ctx, key)
	assert.True(t, acquired, "Should acquire lock after release")

	// Release non-existent key (should not panic)
	l.Release(ctx, "non-existent-key")
}

func TestMemoryLock_MultipleLockWaiters(t *testing.T) {
	t.Parallel()

	l := NewMemoryLock()
	key := "multi-wait"
	ctx := context.Background()

	// Acquire lock
	_, acquired := l.Lock(ctx, key)
	assert.True(t, acquired)

	// Create multiple waiters
	numWaiters := 5
	results := make([]LockResult, numWaiters)
	for i := 0; i < numWaiters; i++ {
		results[i], _ = l.Lock(ctx, key)
		assert.NotNil(t, results[i].ResponseChan)
	}

	// Notify and release - all waiters should receive response
	testResponse := response.NewBuf(200, []byte("test"))
	go l.NotifyAndRelease(ctx, key, testResponse)

	// All waiters should receive the response
	received := 0
	timeout := time.After(2 * time.Second)
	for i := 0; i < numWaiters; i++ {
		select {
		case res := <-results[i].ResponseChan:
			assert.NotNil(t, res)
			assert.Equal(t, 200, res.StatusCode)
			received++
		case <-timeout:
			t.Fatalf("Timeout waiting for waiter %d", i)
		}
	}

	assert.Equal(t, numWaiters, received, "All waiters should receive response")
}

func TestMemoryLock_CancelContext(t *testing.T) {
	t.Parallel()

	l := NewMemoryLock()
	key := "cancel-test"

	ctx, cancel := context.WithCancel(context.Background())

	// Acquire lock
	_, acquired := l.Lock(ctx, key)
	assert.True(t, acquired)

	// Try to acquire with context that will be canceled
	result, acquired := l.Lock(ctx, key)
	assert.False(t, acquired)
	assert.NotNil(t, result.ResponseChan)
	assert.NotNil(t, result.Cancel)

	// Cancel the context
	cancel()

	// Wait a bit for cancel to propagate
	time.Sleep(50 * time.Millisecond)

	// Now notify - the canceled waiter should not receive (channel closed)
	l.NotifyAndRelease(ctx, key, response.NewBuf(200, []byte("test")))

	// Channel should be closed
	_, ok := <-result.ResponseChan
	assert.False(t, ok, "Channel should be closed for canceled context")
}

func TestMemoryLock_Concurrent(t *testing.T) {
	t.Parallel()

	l := NewMemoryLock()
	numGoroutines := 20
	iterations := 10
	key := "concurrent-key"

	done := make(chan bool, numGoroutines)

	for g := 0; g < numGoroutines; g++ {
		go func(id int) {
			ctx := context.Background()
			for i := 0; i < iterations; i++ {
				result, acquired := l.Lock(ctx, key)
				if acquired {
					// Simulate some work
					time.Sleep(1 * time.Millisecond)
					l.NotifyAndRelease(ctx, key, response.NewBuf(200, []byte("ok")))
				} else {
					// Wait for response
					select {
					case res := <-result.ResponseChan:
						assert.NotNil(t, res)
					case <-time.After(1 * time.Second):
						t.Errorf("Goroutine %d iteration %d timed out", id, i)
					}
				}
			}
			done <- true
		}(g)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

func BenchmarkMemoryLock_NotifyAndRelease(b *testing.B) {
	l := NewMemoryLock()
	key := "aaa"
	ctx := context.Background()
	buf := make([]byte, 10)
	l.Lock(ctx, key)
	go time.AfterFunc(time.Millisecond*time.Duration(500), func() {
		l.NotifyAndRelease(ctx, key, response.NewBuf(200, buf))
	})

	for i := 0; i < b.N; i++ {
		result, acquired := l.Lock(ctx, key)
		multi := 500 % (i + 1)
		if acquired {
			go time.AfterFunc(time.Millisecond*time.Duration(multi), func() {
				l.NotifyAndRelease(ctx, key, response.NewBuf(200, buf))
			})
		} else {
			go func(r LockResult) {
				timer := time.NewTimer(time.Second * 1)
				for {
					select {
					case <-r.ResponseChan:
						return
					case <-timer.C:
						panic("timeout waiting for lock")
					default:

					}
				}
			}(result)

			time.Sleep(time.Millisecond * 10)

		}

	}
}
