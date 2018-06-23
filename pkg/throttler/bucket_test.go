package throttler

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewBucketThrottler(t *testing.T) {
	th := NewBucketThrottler(1)

	ctx := context.Background()
	token := th.Take(ctx)

	assert.True(t, token)
}

func TestNewBucketThrottlerBucket(t *testing.T) {
	th := NewBucketThrottlerBacklog(1, 0, time.Millisecond*10)

	ctx := context.Background()
	token := th.Take(ctx)
	assert.True(t, token)

	token = th.Take(ctx)
	assert.False(t, token)

	th.Release()

	token = th.Take(ctx)
	assert.True(t, token)
}
