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
	"github.com/stretchr/testify/assert"
	"net/http"
	"sync"
	"testing"
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
	assert.Equal(t, res.Headers.Get("ETag"), "a588dc2b8c531cd7a1418824963c962d")
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
	assert.Equal(t, res.Headers.Get("ETag"), "3b953319fd6b85711d8074bd70417f4a")
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

	req, _ = http.NewRequest("GET", "http://mort/local/small.jpg-m?width=55", nil)
	res = rp.Process(req, obj)

	assert.Equal(t, res.StatusCode, 200)
	assert.Equal(t, res.Headers.Get("x-amz-meta-public-width"), "100")
	assert.Equal(t, res.Headers.Get("ETag"), "a588dc2b8c531cd7a1418824963c962d")
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

func TestContextTimeout(t *testing.T) {
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

	req, _ = http.NewRequest("HEAD", "http://mort/local/file-test", &buf)
	res = rp.Process(req, obj)

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
