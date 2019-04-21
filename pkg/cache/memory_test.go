package cache

import (
	"testing"
	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/response"
	"github.com/stretchr/testify/assert"
)

func TestMemoryCache_Set(t *testing.T) {
	i := NewMemoryCache(1)

	obj := object.FileObject{}
	obj.Key = "cacheKey"
	res := response.NewString(200, "test")

	i.Set(&obj, res)
	resCache, err := i.Get(&obj)
	b, _ := resCache.ReadBody()
	assert.Nil(t, err)

	assert.Equal(t, resCache.StatusCode, res.StatusCode)
	assert.Equal(t, string(b), "test")
}

func TestMemoryCache_Delete(t *testing.T) {
	i := NewMemoryCache(2)

	obj := object.FileObject{}
	obj.Key = "cacheKey"
	res := response.NewString(200, "test")

	i.Set(&obj, res)
	i.Delete(&obj)
	_, err := i.Get(&obj)
	assert.NotNil(t, err)
}
