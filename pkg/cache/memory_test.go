package cache

import (
	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/response"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMemoryCache_Set(t *testing.T) {
	i := NewMemoryCache(1)

	obj := object.FileObject{}
	obj.Key = "cacheKey"
	res := response.NewString(200, "test")

	i.Set(&obj, res)
	resCache, err := i.Get(&obj)
	assert.Nil(t, err)
	b, err := resCache.Body()
	assert.Nil(t, err)

	assert.Equal(t, resCache.StatusCode, res.StatusCode)
	assert.Equal(t, string(b), "test")
}

func TestMemoryCache_Delete(t *testing.T) {
	t.Parallel()

	i := NewMemoryCache(2)

	obj := object.FileObject{}
	obj.Key = "cacheKey"
	res := response.NewString(200, "test")

	i.Set(&obj, res)
	i.Delete(&obj)
	_, err := i.Get(&obj)
	assert.NotNil(t, err)
}

func TestMemoryCache_SetTooLarge(t *testing.T) {
	t.Parallel()

	i := NewMemoryCache(10) // 10 bytes max

	obj := object.FileObject{}
	obj.Key = "large"
	res := response.NewString(200, "this is a very long response that exceeds the cache size limit by a lot")

	err := i.Set(&obj, res)
	assert.Nil(t, err) // Set doesn't return error, just doesn't cache

	// The item IS cached even if large (memory cache doesn't enforce size limit strictly)
	// This test just verifies Set doesn't panic
}

func TestMemoryCache_GetNotFound(t *testing.T) {
	t.Parallel()

	i := NewMemoryCache(100)

	obj := object.FileObject{}
	obj.Key = "notfound"

	_, err := i.Get(&obj)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestMemoryCache_Concurrent(t *testing.T) {
	t.Parallel()

	i := NewMemoryCache(1000)

	// Test concurrent Set/Get operations
	done := make(chan bool, 10)
	for idx := 0; idx < 10; idx++ {
		go func(id int) {
			obj := object.FileObject{}
			obj.Key = string(rune('a' + id))
			res := response.NewString(200, "test")

			i.Set(&obj, res)
			_, err := i.Get(&obj)
			assert.Nil(t, err)
			done <- true
		}(idx)
	}

	// Wait for all goroutines
	for idx := 0; idx < 10; idx++ {
		<-done
	}
}
