package throttler

import (
	"context"
	"time"
)

// BucketThrottler is implementation of token-bucket algorithm for rate-limiting
type BucketThrottler struct {
	tokens         chan bool
	backlogTokens  chan bool
	backlogTimeout time.Duration
}

// NewBucketThrottler create a new instance of BucketThrottler which limit
func NewBucketThrottler(limit int) *BucketThrottler {
	return NewBucketThrottlerBacklog(limit, 0, defaultBacklogTimeout)
}

// NewBacklog crete a new instance of Throttler which more configuration options
func NewBucketThrottlerBacklog(limit int, backlog int, timeout time.Duration) *BucketThrottler {
	max := limit + backlog
	t := &BucketThrottler{
		tokens:         make(chan bool, limit),
		backlogTokens:  make(chan bool, max),
		backlogTimeout: timeout,
	}

	for i := 0; i < max; i++ {
		if i < limit {
			t.tokens <- true
		}
		t.backlogTokens <- true

	}
	return t
}

// Take retrieve a token from bucket
func (t *BucketThrottler) Take(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	case btok := <-t.backlogTokens:
		timer := time.NewTimer(t.backlogTimeout)

		defer func() {
			t.backlogTokens <- btok
		}()

		select {
		case <-timer.C:
			return false
		case <-t.tokens:
			return true
		}
	default:
		return false
	}
}

// Release return toke to bucket
func (t *BucketThrottler) Release() {
	t.tokens <- true
}
