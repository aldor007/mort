package lock

import (
	"github.com/aldor007/mort/pkg/response"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewNopLock(t *testing.T) {
	l := NewNopLock()

	_, locked := l.Lock("a")
	assert.True(t, locked)

	_, locked = l.Lock("a")
	assert.True(t, locked)

	l.NotifyAndRelease("a", response.NewNoContent(200))

	_, locked = l.Lock("a")
	assert.True(t, locked)

	l.Release("a")

	_, locked = l.Lock("a")
	assert.True(t, locked)
}
