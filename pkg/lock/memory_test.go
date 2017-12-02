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
		t.Fatalf("timeout waitgin for lock")
		return
	case res := <-result.ResponseChan:
		assert.NotNil(t, res, "Response shound't be nil")
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
		buf2, err := res.ReadBody()
		assert.Nil(t, err)
		assert.Equal(t, len(buf), len(buf2))

	}
}

func BenchmarkMemoryLock_NotifyAndRelease(b *testing.B) {
	l := NewMemoryLock()
	key := "aaa"
	buf := make([]byte, 10)

	for i := 0; i < b.N; i++ {
		result, acquired := l.Lock(key)
		multi := 800 % (i + 1)
		if acquired {
			go time.AfterFunc(time.Millisecond*time.Duration(multi), func() {
				go l.NotifyAndRelease(key, response.NewBuf(200, buf))
			})
		} else {

			timer := time.NewTimer(time.Second * 4)
		forLoop:
			for {
				select {
				case <-result.ResponseChan:
					break forLoop
				case <-timer.C:
					b.Fatalf("timeout waitgin for lock")
					return
				default:

				}
			}
		}

	}
}
