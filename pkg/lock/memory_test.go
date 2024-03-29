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
