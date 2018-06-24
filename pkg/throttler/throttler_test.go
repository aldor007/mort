package throttler

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewNopThrottler(t *testing.T) {
	th := NewNopThrottler()
	ctx := context.Background()

	assert.True(t, th.Take(ctx))
	assert.True(t, th.Take(ctx))
}
