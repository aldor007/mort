package lock

import (
	"errors"
	"github.com/aldor007/mort/pkg/response"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewMemoryLock(t *testing.T) {
	l := NewMemoryLock()

	assert.NotNil(t, l, "New memory should return not nil")
}

func TestMemoryLock_Lock(t *testing.T) {
	l := NewMemoryLock()
	key := "klucz"
	c, acquired := l.Lock(key)

	assert.True(t, acquired, "Should acquire lock")
	assert.Nil(t, c.ResponseChan, "shouldn't return channel")

	resChan, lock := l.Lock(key)

	assert.False(t, lock, "Shouldn't acquire lock")
	assert.NotNil(t, resChan, "should return channel")

	l.Release(key)

	c, acquired = l.Lock(key)

	assert.True(t, acquired, "Should acquire lock after release")
	assert.Nil(t, c.ResponseChan, "shouldn't return channel after release")
}

func TestMemoryLock_NotifyAndReleaseWhenError(t *testing.T) {
	l := NewMemoryLock()
	key := "kluczi2"
	c, acquired := l.Lock(key)

	assert.True(t, acquired, "Should acquire lock")
	assert.Nil(t, c.ResponseChan, "shouldn't return channel")

	result, lock := l.Lock(key)

	assert.False(t, lock, "Shouldn't acquire lock")
	assert.NotNil(t, result, "should return channel")

	go l.NotifyAndRelease(key, response.NewError(400, errors.New("invalid transform")))

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
	c, acquired := l.Lock(key)

	assert.True(t, acquired, "Should acquire lock")
	assert.Nil(t, c.ResponseChan, "shouldn't return channel")

	result, lock := l.Lock(key)

	assert.False(t, lock, "Shouldn't acquire lock")
	assert.NotNil(t, result.ResponseChan, "should return channel")

	buf := make([]byte, 1000)
	go l.NotifyAndRelease(key, response.NewBuf(200, buf))

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
	c, acquired := l.Lock(key)

	assert.True(t, acquired, "Should acquire lock")
	assert.Nil(t, c.ResponseChan, "shouldn't return channel")

	l.NotifyAndRelease("no-key", response.NewError(400, errors.New("invalid transform")))
	l.NotifyAndRelease(key, response.NewNoContent(200))
}

func BenchmarkMemoryLock_NotifyAndRelease(b *testing.B) {
	l := NewMemoryLock()
	key := "aaa"
	buf := make([]byte, 10)
	l.Lock(key)
	go time.AfterFunc(time.Millisecond*time.Duration(500), func() {
		l.NotifyAndRelease(key, response.NewBuf(200, buf))
	})

	for i := 0; i < b.N; i++ {
		result, acquired := l.Lock(key)
		multi := 500 % (i + 1)
		if acquired {
			go time.AfterFunc(time.Millisecond*time.Duration(multi), func() {
				l.NotifyAndRelease(key, response.NewBuf(200, buf))
			})
		} else {
			go func(r LockResult) {
				timer := time.NewTimer(time.Second * 1)
				for {
					select {
					case <-r.ResponseChan:
						return
					case <-timer.C:
						b.Fatalf("timeout waiting for lock")
						return
					default:

					}
				}
			}(result)

			time.Sleep(time.Millisecond * 10)

		}

	}
}
