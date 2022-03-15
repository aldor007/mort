package lock

import (
	"context"
	"testing"

	"github.com/aldor007/mort/pkg/response"
	"github.com/stretchr/testify/assert"
)

func TestNewNopLock(t *testing.T) {
	l := NewNopLock()

	ctx := context.Background()
	_, locked := l.Lock(ctx, "a")
	assert.True(t, locked)

	_, locked = l.Lock(ctx, "a")
	assert.True(t, locked)

	l.NotifyAndRelease(ctx, "a", response.NewNoContent(200))

	_, locked = l.Lock(ctx, "a")
	assert.True(t, locked)

	l.Release(ctx, "a")

	_, locked = l.Lock(ctx, "a")
	assert.True(t, locked)
}
