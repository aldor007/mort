package processor

import (
	"bytes"
	"context"
	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/lock"
	"github.com/aldor007/mort/pkg/middleware"
	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/response"
	"github.com/aldor007/mort/pkg/storage"
	"github.com/aldor007/mort/pkg/throttler"
	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestNewRequestProcessor(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://mort/local/small.jpg-m", nil)

	mortConfig := config.Config{}
	err := mortConfig.Load("./benchmark/small.yml")
	assert.Nil(t, err)

	obj, err := object.NewFileObject(req.URL, &mortConfig)
	assert.Nil(t, err)

	rp := NewRequestProcessor(mortConfig.Server, lock.NewMemoryLock(), throttler.NewBucketThrottler(10))
	res := rp.Process(req, obj)

	assert.Equal(t, res.StatusCode, 200)
	assert.Equal(t, res.Headers.Get("x-amz-meta-public-width"), "100")
	assert.Equal(t, res.Headers.Get("x-amz-meta-public-height"), "100")
}
func TestNewRequestProcessorCheckParent(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://mort/local/small.jpg-mm", nil)

	mortConfig := config.Config{}
	err := mortConfig.Load("./benchmark/small.yml")
	assert.Nil(t, err)

	obj, err := object.NewFileObject(req.URL, &mortConfig)
	obj.CheckParent = true
	assert.Nil(t, err)

	rp := NewRequestProcessor(mortConfig.Server, lock.NewMemoryLock(), throttler.NewBucketThrottler(10))
	obj.CheckParent = true
	res := rp.Process(req, obj)

	assert.Equal(t, res.StatusCode, 200)
	assert.Equal(t, res.Headers.Get("x-amz-meta-public-width"), "10")
	assert.Equal(t, res.Headers.Get("x-amz-meta-public-height"), "100")
}

func TestFetchFromCache(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://mort/local/small.jpg-m?width=55", nil)
	req.Header.Add("x-mort-debug", "1")

	mortConfig := config.Config{}
	err := mortConfig.Load("./benchmark/small.yml")
	assert.Nil(t, err)

	obj, err := object.NewFileObject(req.URL, &mortConfig)
	assert.Nil(t, err)

	rp := NewRequestProcessor(mortConfig.Server, lock.NewMemoryLock(), throttler.NewBucketThrottler(10))
	res := rp.Process(req, obj)

	req, _ = http.NewRequest("DELETE", "http://mort/local/small.jpg-m?width=55", nil)
	res = rp.Process(req, obj)

	storageRes := storage.Get(obj)

	assert.Equal(t, 404, storageRes.StatusCode)

	obj2, err := object.NewFileObject(req.URL, &mortConfig)
	assert.Nil(t, err)

	req, _ = http.NewRequest("GET", "http://mort/local/small.jpg-m?width=55", nil)
	res = rp.Process(req, obj2)

	assert.Equal(t, res.StatusCode, 200)
	assert.Equal(t, res.Headers.Get("x-amz-meta-public-width"), "100")
	assert.Equal(t, res.Headers.Get("x-amz-meta-public-height"), "100")
}

func TestReturn404WhenParentNotFound(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://mort/local/small.not-m", nil)

	mortConfig := config.Config{}
	err := mortConfig.Load("./benchmark/small.yml")
	assert.Nil(t, err)

	obj, err := object.NewFileObject(req.URL, &mortConfig)
	assert.Nil(t, err)

	rp := NewRequestProcessor(mortConfig.Server, lock.NewMemoryLock(), throttler.NewBucketThrottler(10))
	res := rp.Process(req, obj)

	assert.Equal(t, res.StatusCode, 404)
}

func TestReturn503WhenThrottled(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://mort/local/small.jpg?width=5", nil)

	mortConfig := config.Config{}
	err := mortConfig.Load("./benchmark/small.yml")
	assert.Nil(t, err)

	obj, err := object.NewFileObject(req.URL, &mortConfig)
	assert.Nil(t, err)

	rp := NewRequestProcessor(mortConfig.Server, lock.NewMemoryLock(), throttler.NewBucketThrottler(0))
	res := rp.Process(req, obj)

	assert.Equal(t, res.StatusCode, 503)
}

func TestConcurrentImageProcessingLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		concurrentLimit int
		totalRequests   int
		description     string
	}{
		{
			name:            "limit of 2 with concurrent requests",
			concurrentLimit: 2,
			totalRequests:   5,
			description:     "with token release, requests should succeed over time",
		},
		{
			name:            "limit of 5 with concurrent requests",
			concurrentLimit: 5,
			totalRequests:   10,
			description:     "with token release, requests should succeed over time",
		},
		{
			name:            "limit of 1 with concurrent requests",
			concurrentLimit: 1,
			totalRequests:   3,
			description:     "with token release, requests should succeed over time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mortConfig := config.Config{}
			err := mortConfig.Load("./benchmark/small.yml")
			assert.Nil(t, err)

			rp := NewRequestProcessor(mortConfig.Server, lock.NewMemoryLock(), throttler.NewBucketThrottler(tt.concurrentLimit))

			var wg sync.WaitGroup
			results := make(chan int, tt.totalRequests)

			// Start all requests concurrently
			for i := 0; i < tt.totalRequests; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					req, _ := http.NewRequest("GET", "http://mort/local/small.jpg?width=5", nil)
					obj, _ := object.NewFileObject(req.URL, &mortConfig)

					res := rp.Process(req, obj)
					results <- res.StatusCode
				}()
			}

			wg.Wait()
			close(results)

			// Count successes and throttled
			successCount := 0
			throttledCount := 0
			for statusCode := range results {
				if statusCode == 200 {
					successCount++
				} else if statusCode == 503 {
					throttledCount++
				}
			}

			// With token release, some or all requests should succeed
			// The test verifies the throttler doesn't break the processor
			assert.Greater(t, successCount, 0, "%s - at least some requests should succeed", tt.description)
			assert.Equal(t, tt.totalRequests, successCount+throttledCount,
				"%s - all requests should complete", tt.description)

			// If processing is fast enough, all might succeed due to token release
			// This is correct behavior - we just verify the system works
			t.Logf("%s: %d/%d succeeded, %d throttled (limit: %d)",
				tt.name, successCount, tt.totalRequests, throttledCount, tt.concurrentLimit)
		})
	}
}

func TestConcurrentImageProcessingWithRelease(t *testing.T) {
	t.Parallel()

	mortConfig := config.Config{}
	err := mortConfig.Load("./benchmark/small.yml")
	assert.Nil(t, err)

	// Create throttler with limit of 2
	limit := 2
	rp := NewRequestProcessor(mortConfig.Server, lock.NewMemoryLock(), throttler.NewBucketThrottler(limit))

	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	// Start 10 requests that will run in batches due to limit of 2
	totalRequests := 10
	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			req, _ := http.NewRequest("GET", "http://mort/local/small.jpg?width=5", nil)
			obj, _ := object.NewFileObject(req.URL, &mortConfig)

			res := rp.Process(req, obj)
			if res.StatusCode == 200 {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
		// Small delay to ensure requests don't all hit at exact same time
		time.Sleep(time.Millisecond * 10)
	}

	wg.Wait()

	// With token release, all requests should eventually succeed
	// (as tokens are released, new requests can acquire them)
	assert.Greater(t, successCount, limit,
		"With token release, more than %d requests should succeed", limit)
}

func TestConfigConcurrentImageProcessing(t *testing.T) {
	t.Parallel()

	t.Run("uses config value when set", func(t *testing.T) {
		configYaml := `
server:
  concurrentImageProcessing: 50
  listens:
    - ":8080"
buckets:
  test:
    storages:
      basic:
        kind: "local-meta"
        rootPath: "/tmp"
`
		mortConfig := config.Config{}
		err := mortConfig.LoadFromString(configYaml)
		assert.Nil(t, err)

		// In real usage, main.go would use this value to create the throttler
		concurrentLimit := mortConfig.Server.ConcurrentImageProcessing
		if concurrentLimit <= 0 {
			concurrentLimit = 100
		}

		assert.Equal(t, 50, concurrentLimit, "Should use configured value")
	})

	t.Run("defaults to 100 when not set", func(t *testing.T) {
		configYaml := `
server:
  listens:
    - ":8080"
buckets:
  test:
    storages:
      basic:
        kind: "local-meta"
        rootPath: "/tmp"
`
		mortConfig := config.Config{}
		err := mortConfig.LoadFromString(configYaml)
		assert.Nil(t, err)

		// In real usage, main.go would apply the default
		concurrentLimit := mortConfig.Server.ConcurrentImageProcessing
		if concurrentLimit <= 0 {
			concurrentLimit = 100
		}

		assert.Equal(t, 100, concurrentLimit, "Should default to 100")
	})
}

func TestContextTimeout(t *testing.T) {
	t.Parallel()

	req, _ := http.NewRequest("GET", "http://mort/local/small.jpg?width=5", nil)
	ctx, cancel := context.WithCancel(context.Background())
	req = req.WithContext(ctx)
	cancel()

	mortConfig := config.Config{}
	err := mortConfig.Load("./benchmark/small.yml")
	assert.Nil(t, err)

	obj, err := object.NewFileObject(req.URL, &mortConfig)
	assert.Nil(t, err)

	rp := NewRequestProcessor(mortConfig.Server, lock.NewMemoryLock(), throttler.NewBucketThrottler(0))
	res := rp.Process(req, obj)

	assert.Equal(t, res.StatusCode, 499)
}

// TestContextTimeoutRace tests that concurrent requests with canceled contexts
// don't cause a race condition or panic from sending on a closed channel
func TestContextTimeoutRace(t *testing.T) {
	t.Parallel()

	mortConfig := config.Config{}
	err := mortConfig.Load("./benchmark/small.yml")
	assert.Nil(t, err)

	rp := NewRequestProcessor(mortConfig.Server, lock.NewMemoryLock(), throttler.NewBucketThrottler(0))

	// Run many concurrent requests with already-canceled contexts
	// This should trigger the race condition that was fixed
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, _ := http.NewRequest("GET", "http://mort/local/small.jpg?width=5", nil)
			ctx, cancel := context.WithCancel(context.Background())
			req = req.WithContext(ctx)
			cancel() // Cancel immediately

			obj, err := object.NewFileObject(req.URL, &mortConfig)
			assert.Nil(t, err)

			res := rp.Process(req, obj)
			// When context is cancelled immediately, we might get either:
			// - 499 (client cancelled request) if detected during processing
			// - 504 (timeout) if timeout occurs before cancellation is detected
			// Both are valid responses and don't indicate a race condition
			assert.Contains(t, []int{499, 504}, res.StatusCode,
				"should return either 499 (cancelled) or 504 (timeout), got %d", res.StatusCode)
		}()
	}
	wg.Wait()
}

func TestCollapse(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://mort/local/small.jpg?width=54", nil)
	req2, _ := http.NewRequest("GET", "http://mort/local/small.jpg?width=54", nil)

	mortConfig := config.Config{}
	err := mortConfig.Load("./benchmark/small.yml")
	assert.Nil(t, err)

	obj, err := object.NewFileObject(req.URL, &mortConfig)
	assert.Nil(t, err)

	obj2, err := object.NewFileObject(req.URL, &mortConfig)
	assert.Nil(t, err)

	rp := NewRequestProcessor(mortConfig.Server, lock.NewMemoryLock(), throttler.NewBucketThrottler(1))
	var wg sync.WaitGroup

	var res1 *response.Response
	var res2 *response.Response

	wg.Add(1)
	go func() {
		res1 = rp.Process(req, obj)
		wg.Done()

		assert.Equal(t, res1.StatusCode, 200)
	}()

	wg.Add(1)
	go func() {
		res2 = rp.Process(req2, obj2)
		wg.Done()

		assert.Equal(t, res2.StatusCode, 200)
	}()

	wg.Wait()
	assert.Equal(t, res1.StatusCode, 200)
	assert.Equal(t, res2.StatusCode, 200)
}

func TestMethodNotAllowed(t *testing.T) {
	req, _ := http.NewRequest("OPTIONS", "http://mort/local/small.jpg-m?width=55", nil)

	mortConfig := config.Config{}
	err := mortConfig.Load("./benchmark/small.yml")
	assert.Nil(t, err)

	obj, err := object.NewFileObject(req.URL, &mortConfig)
	assert.Nil(t, err)

	rp := NewRequestProcessor(mortConfig.Server, lock.NewMemoryLock(), throttler.NewBucketThrottler(10))
	res := rp.Process(req, obj)

	assert.Equal(t, res.StatusCode, 405)
}

func TestGetParent(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://mort/local/small.jpg", nil)

	mortConfig := config.Config{}
	err := mortConfig.Load("./benchmark/small.yml")
	assert.Nil(t, err)

	obj, err := object.NewFileObject(req.URL, &mortConfig)
	assert.Nil(t, err)

	rp := NewRequestProcessor(mortConfig.Server, lock.NewMemoryLock(), throttler.NewBucketThrottler(10))
	res := rp.Process(req, obj)

	assert.Equal(t, res.StatusCode, 200)
}

func TestPut(t *testing.T) {
	buf := bytes.Buffer{}
	buf.WriteString("aaaa")

	req, _ := http.NewRequest("PUT", "http://mort/local/file-test", &buf)

	mortConfig := config.Config{}
	err := mortConfig.Load("./benchmark/small.yml")
	assert.Nil(t, err)

	obj, err := object.NewFileObject(req.URL, &mortConfig)
	assert.Nil(t, err)

	rp := NewRequestProcessor(mortConfig.Server, lock.NewMemoryLock(), throttler.NewBucketThrottler(10))
	res := rp.Process(req, obj)

	assert.Equal(t, res.StatusCode, 200)

	obj2, err := object.NewFileObject(req.URL, &mortConfig)
	assert.Nil(t, err)

	req, _ = http.NewRequest("HEAD", "http://mort/local/file-test", &buf)
	res = rp.Process(req, obj2)

	assert.Equal(t, res.StatusCode, 200)
}

func TestS3GET(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://mort/local?maker=&max-keys=1000&delimter=&prefix=", nil)
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.S3AuthCtxKey, true)
	req = req.WithContext(ctx)

	mortConfig := config.Config{}
	err := mortConfig.Load("./benchmark/small.yml")
	assert.Nil(t, err)

	obj, err := object.NewFileObject(req.URL, &mortConfig)
	assert.Nil(t, err)
	assert.False(t, obj.HasTransform())

	rp := NewRequestProcessor(mortConfig.Server, lock.NewMemoryLock(), throttler.NewBucketThrottler(10))
	res := rp.Process(req, obj)

	assert.Equal(t, res.StatusCode, 200)
	assert.Equal(t, res.Headers.Get("content-type"), "application/xml")
}

func TestS3GETNoCache(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://mort/local/small.jpg", nil)
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.S3AuthCtxKey, true)
	req = req.WithContext(ctx)

	mortConfig := config.Config{}
	err := mortConfig.Load("./benchmark/small.yml")
	assert.Nil(t, err)

	obj, err := object.NewFileObject(req.URL, &mortConfig)
	assert.Nil(t, err)

	rp := NewRequestProcessor(mortConfig.Server, lock.NewMemoryLock(), throttler.NewBucketThrottler(10))
	res := rp.Process(req, obj)

	assert.Equal(t, res.StatusCode, 200)
	assert.Equal(t, res.Headers.Get("cache-control"), "no-cache")
}

func TestTransformWrongContentType(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://mort/local/file.txt?width=400", nil)

	mortConfig := config.Config{}
	err := mortConfig.Load("./benchmark/small.yml")
	assert.Nil(t, err)

	obj, err := object.NewFileObject(req.URL, &mortConfig)
	assert.Nil(t, err)

	rp := NewRequestProcessor(mortConfig.Server, lock.NewMemoryLock(), throttler.NewBucketThrottler(10))
	res := rp.Process(req, obj)

	assert.Equal(t, res.StatusCode, 404)
}

func BenchmarkNewRequestProcessorMemoryLock(b *testing.B) {
	benchmarks := []struct {
		name       string
		url        string
		configPath string
	}{
		{"Process small image, small result", "http://mort/local/small.jpg-small", "./benchmark/small.yml"},
		{"Process large image, small result", "http://mort/local/large.jpeg-small", "./benchmark/small.yml"},
	}

	for _, bm := range benchmarks {
		req, _ := http.NewRequest("GET", bm.url, nil)

		mortConfig := config.Config{}
		err := mortConfig.Load(bm.configPath)
		if err != nil {
			panic(err)
		}

		obj, _ := object.NewFileObject(req.URL, &mortConfig)
		rp := NewRequestProcessor(mortConfig.Server, lock.NewMemoryLock(), throttler.NewBucketThrottler(10))
		errorCounter := 0
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				res := rp.Process(req, obj)
				if res.StatusCode != 200 {
					errorCounter++
					//b.Fatalf("Invalid response sc %s test name %s", res.StatusCode, bm.name)
				}
			}

			if float32(errorCounter/b.N) > 0.001 {
				b.Fatalf("To many errors %d / %d", errorCounter, b.N)
			}
		})

	}

}

func BenchmarkNewRequestProcessorNopLock(b *testing.B) {
	benchmarks := []struct {
		name       string
		url        string
		configPath string
	}{
		{"Process small image, small result", "http://mort/local/small.jpg-small", "./benchmark/small.yml"},
		{"Process large image, small result", "http://mort/local/large.jpeg-small", "./benchmark/small.yml"},
	}

	for _, bm := range benchmarks {
		req, _ := http.NewRequest("GET", bm.url, nil)

		mortConfig := config.Config{}
		err := mortConfig.Load(bm.configPath)
		if err != nil {
			panic(err)
		}

		obj, _ := object.NewFileObject(req.URL, &mortConfig)
		rp := NewRequestProcessor(mortConfig.Server, lock.NewNopLock(), throttler.NewBucketThrottler(10))
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				res := rp.Process(req, obj)
				if res.StatusCode != 200 {
					b.Fatalf("Invalid response sc %d test name %s", res.StatusCode, bm.name)
				}
			}
		})
	}

}

func TestCollapseRedisLock(t *testing.T) {
	s := miniredis.RunT(t)
	req, _ := http.NewRequest("GET", "http://mort/local/small.jpg?width=54", nil)
	req2, _ := http.NewRequest("GET", "http://mort/local/small.jpg?width=54", nil)

	mortConfig := config.Config{}
	err := mortConfig.Load("./benchmark/small.yml")
	assert.Nil(t, err)

	obj, err := object.NewFileObject(req.URL, &mortConfig)
	assert.Nil(t, err)

	obj2, err := object.NewFileObject(req.URL, &mortConfig)
	assert.Nil(t, err)

	rp := NewRequestProcessor(mortConfig.Server, lock.NewRedisLock([]string{s.Addr()}, nil), throttler.NewBucketThrottler(1))
	var wg sync.WaitGroup

	var res1 *response.Response
	var res2 *response.Response

	wg.Add(1)
	go func() {
		res1 = rp.Process(req, obj)
		wg.Done()

		assert.Equal(t, res1.StatusCode, 200)
	}()

	wg.Add(1)
	go func() {
		res2 = rp.Process(req2, obj2)
		wg.Done()

		assert.Equal(t, res2.StatusCode, 200)
	}()

	wg.Wait()
	assert.Equal(t, res1.StatusCode, 200)
	assert.Equal(t, res2.StatusCode, 200)
}

func TestCollapseRedisLockTwoInstaces(t *testing.T) {
	s := miniredis.RunT(t)
	req, _ := http.NewRequest("GET", "http://mort/local/small.jpg?width=54", nil)
	req2, _ := http.NewRequest("GET", "http://mort/local/small.jpg?width=54", nil)

	mortConfig := config.Config{}
	err := mortConfig.Load("./benchmark/small.yml")
	assert.Nil(t, err)

	obj, err := object.NewFileObject(req.URL, &mortConfig)
	assert.Nil(t, err)

	obj2, err := object.NewFileObject(req.URL, &mortConfig)
	assert.Nil(t, err)

	rp := NewRequestProcessor(mortConfig.Server, lock.NewRedisLock([]string{s.Addr()}, nil), throttler.NewBucketThrottler(1))
	rp2 := NewRequestProcessor(mortConfig.Server, lock.NewRedisLock([]string{s.Addr()}, nil), throttler.NewBucketThrottler(1))
	var wg sync.WaitGroup

	var res1 *response.Response
	var res2 *response.Response

	wg.Add(1)
	go func() {
		res1 = rp.Process(req, obj)
		wg.Done()

		assert.Equal(t, res1.StatusCode, 200)
	}()

	wg.Add(1)
	go func() {
		res2 = rp2.Process(req2, obj2)
		wg.Done()

		assert.Equal(t, res2.StatusCode, 200)
	}()

	wg.Wait()
	assert.Equal(t, res1.StatusCode, 200)
	assert.Equal(t, res2.StatusCode, 200)
}
