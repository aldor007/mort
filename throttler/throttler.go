package throttler

import (
	"time"
	"context"
)


// defaultBacklogTimeout set to 60s
var defaultBacklogTimeout = time.Second * 60

// Throttler is rate limiter
type Throttler interface {
	Take(ctx context.Context) (taken bool) // Take tries acquire token when its true its mean you can process when false have been throttled
	Release() // Release returns token to pool
}

// NopThrottler is always return that you can perform given operation
type NopThrottler struct {

}

// NewNopThrottler create instance of NopThrottler
func NewNopThrottler(_ ...interface{}) *NopThrottler {
	return &NopThrottler{}
}

func (*NopThrottler) Take(_ context.Context) bool {
	return true;
}

func (*NopThrottler) Release() {

}
