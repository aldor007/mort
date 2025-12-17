package response

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSharedResponse(t *testing.T) {
	t.Parallel()

	t.Run("should create SharedResponse from buffered response", func(t *testing.T) {
		res := NewBuf(200, []byte("test data"))
		shared, err := NewSharedResponse(res)

		assert.Nil(t, err, "should not return error for buffered response")
		assert.NotNil(t, shared, "should return SharedResponse")
		assert.Equal(t, int32(1), shared.RefCount(), "initial refcount should be 1")
	})

	t.Run("should fail with nil response", func(t *testing.T) {
		_, err := NewSharedResponse(nil)

		assert.NotNil(t, err, "should return error for nil response")
	})

	t.Run("should work with already buffered response", func(t *testing.T) {
		// Create a properly buffered response
		res := NewString(200, "test")

		shared, err := NewSharedResponse(res)
		assert.Nil(t, err, "should not return error for buffered response")
		assert.NotNil(t, shared, "should return SharedResponse")

		view := shared.Acquire()
		body, _ := view.Body()
		assert.Equal(t, "test", string(body), "should have correct body")
		shared.Release()
		shared.Release() // Release original reference
	})
}

func TestSharedResponse_Acquire(t *testing.T) {
	t.Parallel()

	t.Run("should increment refcount on acquire", func(t *testing.T) {
		res := NewBuf(200, []byte("test data"))
		shared, _ := NewSharedResponse(res)

		view1 := shared.Acquire()
		assert.Equal(t, int32(2), shared.RefCount(), "refcount should be 2 after first acquire")

		view2 := shared.Acquire()
		assert.Equal(t, int32(3), shared.RefCount(), "refcount should be 3 after second acquire")

		// Verify views have correct data
		body1, _ := view1.Body()
		body2, _ := view2.Body()
		assert.Equal(t, "test data", string(body1), "view1 should have correct body")
		assert.Equal(t, "test data", string(body2), "view2 should have correct body")
	})

	t.Run("acquired views should have same status code and headers", func(t *testing.T) {
		res := NewBuf(200, []byte("test"))
		res.Set("X-Test-Header", "test-value")
		shared, _ := NewSharedResponse(res)

		view := shared.Acquire()

		assert.Equal(t, 200, view.StatusCode, "view should have same status code")
		assert.Equal(t, "test-value", view.Headers.Get("X-Test-Header"), "view should have headers")
	})

	t.Run("should share body buffer between views", func(t *testing.T) {
		originalData := []byte("shared data")
		res := NewBuf(200, originalData)
		shared, _ := NewSharedResponse(res)

		view1 := shared.Acquire()
		view2 := shared.Acquire()

		body1, _ := view1.Body()
		body2, _ := view2.Body()

		// Verify both views see the same data (buffer is shared)
		assert.Equal(t, originalData, body1, "view1 should have shared data")
		assert.Equal(t, originalData, body2, "view2 should have shared data")
	})
}

func TestSharedResponse_Release(t *testing.T) {
	t.Parallel()

	t.Run("should decrement refcount on release", func(t *testing.T) {
		res := NewBuf(200, []byte("test"))
		shared, _ := NewSharedResponse(res)

		shared.Acquire()
		assert.Equal(t, int32(2), shared.RefCount(), "refcount should be 2")

		shared.Release()
		assert.Equal(t, int32(1), shared.RefCount(), "refcount should be 1 after release")

		shared.Release()
		assert.Equal(t, int32(0), shared.RefCount(), "refcount should be 0 after final release")
	})

	t.Run("should not panic on multiple releases", func(t *testing.T) {
		res := NewBuf(200, []byte("test"))
		shared, _ := NewSharedResponse(res)

		shared.Release() // Release initial reference
		assert.NotPanics(t, func() {
			shared.Release() // Additional release should not panic (refcount goes negative)
		})
	})
}

func TestSharedResponse_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	t.Run("should handle concurrent acquire/release safely", func(t *testing.T) {
		res := NewBuf(200, []byte("concurrent test data"))
		shared, _ := NewSharedResponse(res)

		numGoroutines := 100
		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		// Concurrent acquire and release
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				view := shared.Acquire()
				body, err := view.Body()
				assert.Nil(t, err, "should read body without error")
				assert.Equal(t, "concurrent test data", string(body), "should have correct body")
				shared.Release()
			}()
		}

		wg.Wait()

		// After all goroutines finish, refcount should be back to 1 (original reference)
		shared.Release() // Release original reference
		assert.Equal(t, int32(0), shared.RefCount(), "refcount should be 0 after all releases")
	})

	t.Run("should safely read from multiple views concurrently", func(t *testing.T) {
		testData := make([]byte, 1024*10) // 10KB test data
		for i := range testData {
			testData[i] = byte(i % 256)
		}
		res := NewBuf(200, testData)
		shared, _ := NewSharedResponse(res)

		numReaders := 50
		var wg sync.WaitGroup
		wg.Add(numReaders)

		for i := 0; i < numReaders; i++ {
			go func() {
				defer wg.Done()
				view := shared.Acquire()
				defer shared.Release()

				body, err := view.Body()
				assert.Nil(t, err, "should read body without error")
				assert.Equal(t, len(testData), len(body), "should have correct body length")
				assert.Equal(t, testData, body, "should have identical body content")
			}()
		}

		wg.Wait()
		shared.Release() // Release original reference
	})
}

func TestSharedResponse_MemoryOptimization(t *testing.T) {
	t.Parallel()

	t.Run("should not duplicate large buffers", func(t *testing.T) {
		// Create a 5MB buffer
		largeData := make([]byte, 5*1024*1024)
		res := NewBuf(200, largeData)
		shared, _ := NewSharedResponse(res)

		// Acquire 10 views
		views := make([]*Response, 10)
		for i := 0; i < 10; i++ {
			views[i] = shared.Acquire()
		}

		// Verify all views reference the same buffer (no copies)
		// This is implicit - we can't directly test pointer equality through the API,
		// but we can verify the data is correct and the operation is fast
		for _, view := range views {
			body, _ := view.Body()
			assert.Equal(t, len(largeData), len(body), "view should have correct size")
		}

		// Clean up
		for i := 0; i < 10; i++ {
			shared.Release()
		}
		shared.Release() // original reference
	})
}
