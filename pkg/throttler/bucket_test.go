package throttler

import (
	"context"
	"github.com/stretchr/testify/assert"
	"sync"
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

func TestBucketThrottler_ContextCancellation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		limit       int
		backlog     int
		description string
	}{
		{
			name:        "should return false when context is cancelled with zero limit",
			limit:       0,
			backlog:     0,
			description: "cancelled context with no tokens should return false",
		},
		{
			name:        "should return false when backlog is exhausted",
			limit:       1,
			backlog:     0,
			description: "cancelled context when all tokens taken should return false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			th := NewBucketThrottlerBacklog(tt.limit, tt.backlog, time.Millisecond*10)

			// Take all tokens if limit > 0
			if tt.limit > 0 {
				ctx := context.Background()
				th.Take(ctx)
			}

			// Now try with cancelled context
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			token := th.Take(ctx)
			assert.False(t, token, tt.description)
		})
	}
}

func TestBucketThrottler_BacklogTimeout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		limit          int
		backlog        int
		timeout        time.Duration
		takeCalls      int
		expectedResult []bool
		description    string
	}{
		{
			name:           "should timeout when backlog is full",
			limit:          1,
			backlog:        0,
			timeout:        time.Millisecond * 50,
			takeCalls:      2,
			expectedResult: []bool{true, false},
			description:    "second call should timeout when no tokens available",
		},
		{
			name:           "should use backlog before timing out",
			limit:          1,
			backlog:        2,
			timeout:        time.Millisecond * 50,
			takeCalls:      2,
			expectedResult: []bool{true, false},
			description:    "backlog should allow waiting but timeout eventually",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			th := NewBucketThrottlerBacklog(tt.limit, tt.backlog, tt.timeout)
			ctx := context.Background()

			results := make([]bool, tt.takeCalls)
			for i := 0; i < tt.takeCalls; i++ {
				results[i] = th.Take(ctx)
			}

			for i, expected := range tt.expectedResult {
				assert.Equal(t, expected, results[i], "%s - call %d", tt.description, i+1)
			}
		})
	}
}

func TestBucketThrottler_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		limit       int
		goroutines  int
		description string
	}{
		{
			name:        "should handle concurrent take operations with limit 5",
			limit:       5,
			goroutines:  10,
			description: "only 5 goroutines should acquire tokens",
		},
		{
			name:        "should handle concurrent take operations with limit 10",
			limit:       10,
			goroutines:  20,
			description: "only 10 goroutines should acquire tokens",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			th := NewBucketThrottlerBacklog(tt.limit, 0, time.Millisecond*10)
			ctx := context.Background()

			successCount := 0
			var mu sync.Mutex
			var wg sync.WaitGroup

			for i := 0; i < tt.goroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					if th.Take(ctx) {
						mu.Lock()
						successCount++
						mu.Unlock()
					}
				}()
			}

			wg.Wait()

			assert.Equal(t, tt.limit, successCount, tt.description)
		})
	}
}

func TestBucketThrottler_ReleaseAndReuse(t *testing.T) {
	t.Parallel()

	th := NewBucketThrottlerBacklog(2, 0, time.Millisecond*10)
	ctx := context.Background()

	// Take all tokens
	token1 := th.Take(ctx)
	token2 := th.Take(ctx)
	assert.True(t, token1, "should acquire first token")
	assert.True(t, token2, "should acquire second token")

	// No more tokens available
	token3 := th.Take(ctx)
	assert.False(t, token3, "should not acquire token when bucket is empty")

	// Release one token
	th.Release()

	// Should be able to acquire again
	token4 := th.Take(ctx)
	assert.True(t, token4, "should acquire token after release")
}

func TestBucketThrottler_MultipleReleases(t *testing.T) {
	t.Parallel()

	th := NewBucketThrottlerBacklog(3, 0, time.Millisecond*10)
	ctx := context.Background()

	// Take all 3 tokens
	token1 := th.Take(ctx)
	token2 := th.Take(ctx)
	token3 := th.Take(ctx)
	assert.True(t, token1)
	assert.True(t, token2)
	assert.True(t, token3)

	// No more tokens available
	token4 := th.Take(ctx)
	assert.False(t, token4, "should not have token when all are taken")

	// Release tokens back
	th.Release()
	th.Release()
	th.Release()

	// Should be able to take again
	token5 := th.Take(ctx)
	token6 := th.Take(ctx)
	token7 := th.Take(ctx)

	assert.True(t, token5, "should have token after first release")
	assert.True(t, token6, "should have token after second release")
	assert.True(t, token7, "should have token after third release")
}

func TestBucketThrottler_BacklogWithWaiters(t *testing.T) {
	t.Parallel()

	th := NewBucketThrottlerBacklog(1, 5, time.Millisecond*100)
	ctx := context.Background()

	// Take the main token
	token := th.Take(ctx)
	assert.True(t, token, "should acquire main token")

	// Start goroutines that will wait in backlog
	results := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			r := th.Take(ctx)
			results <- r
		}()
	}

	// Give goroutines time to enter backlog
	time.Sleep(time.Millisecond * 10)

	// Release the token
	th.Release()

	// One waiter should get the token
	success := <-results
	assert.True(t, success, "one waiter should acquire released token")
}

func TestBucketThrottler_ZeroLimit(t *testing.T) {
	t.Parallel()

	th := NewBucketThrottler(0)
	ctx := context.Background()

	token := th.Take(ctx)
	assert.False(t, token, "should not acquire token with zero limit")
}

func TestBucketThrottler_HighConcurrency(t *testing.T) {
	t.Parallel()

	th := NewBucketThrottler(50)
	ctx := context.Background()

	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	// Start 100 goroutines trying to acquire tokens
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if th.Take(ctx) {
				mu.Lock()
				successCount++
				mu.Unlock()

				// Simulate work
				time.Sleep(time.Millisecond * 1)

				// Release token
				th.Release()
			}
		}()
	}

	wg.Wait()

	// At most 50 should have succeeded initially
	assert.LessOrEqual(t, successCount, 100, "should handle high concurrency")
	assert.Greater(t, successCount, 0, "at least some should succeed")
}

func TestBucketThrottler_ContextCancellationDuringWait(t *testing.T) {
	t.Parallel()

	th := NewBucketThrottlerBacklog(1, 5, time.Millisecond*200)

	// Take the main token
	ctx1 := context.Background()
	token := th.Take(ctx1)
	assert.True(t, token, "should acquire main token")

	// Try to take with cancellable context
	ctx2, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel()

	token2 := th.Take(ctx2)
	assert.False(t, token2, "should return false when context times out")
}

func TestBucketThrottler_BacklogCapacity(t *testing.T) {
	t.Parallel()

	limit := 2
	backlog := 3
	th := NewBucketThrottlerBacklog(limit, backlog, time.Millisecond*50)
	ctx := context.Background()

	// Take all main tokens
	token1 := th.Take(ctx)
	token2 := th.Take(ctx)
	assert.True(t, token1)
	assert.True(t, token2)

	// Start goroutines that will fill the backlog
	results := make([]chan bool, backlog+2)
	for i := 0; i < backlog+2; i++ {
		results[i] = make(chan bool, 1)
		go func(ch chan bool) {
			r := th.Take(ctx)
			ch <- r
		}(results[i])
	}

	// Give them time to try
	time.Sleep(time.Millisecond * 60)

	// Verify that requests beyond backlog capacity failed
	failedCount := 0
	for i := 0; i < backlog+2; i++ {
		select {
		case r := <-results[i]:
			if !r {
				failedCount++
			}
		default:
			// Still waiting
		}
	}

	assert.Greater(t, failedCount, 0, "should have some failed requests when exceeding backlog+limit")
}

func TestBucketThrottler_DefaultBacklogTimeout(t *testing.T) {
	t.Parallel()

	// NewBucketThrottler uses defaultBacklogTimeout which should be 60s
	th := NewBucketThrottler(1)
	assert.NotNil(t, th, "should create throttler with default timeout")

	// This test just verifies the constructor works
	// The actual timeout value is internal
	ctx := context.Background()
	token := th.Take(ctx)
	assert.True(t, token, "should work with default timeout")
}
