package processor

import (
	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/lock"
	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/throttler"
	//"github.com/aldor007/mort/pkg/monitoring"
	//"go.uber.org/zap"
	"bytes"
	"context"
	"github.com/aldor007/mort/pkg/storage"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"github.com/aldor007/mort/pkg/storage"
	"bytes"
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
	assert.Equal(t, res.Headers.Get("x-amz-meta-public-width"), "150")
	assert.Equal(t, res.Headers.Get("ETag"), "W/\"7eaa484e8c841e7e\"")
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
	assert.Equal(t, res.Headers.Get("x-amz-meta-public-width"), "150")
	assert.Equal(t, res.Headers.Get("ETag"), "W/\"7eaa484e8c841e7e\"")
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

	req, _ := http.NewRequest("PUT", "http://mort/local/fila-test", &buf)

	mortConfig := config.Config{}
	err := mortConfig.Load("./benchmark/small.yml")
	assert.Nil(t, err)

	obj, err := object.NewFileObject(req.URL, &mortConfig)
	assert.Nil(t, err)

	rp := NewRequestProcessor(mortConfig.Server, lock.NewMemoryLock(), throttler.NewBucketThrottler(10))
	res := rp.Process(req, obj)

	assert.Equal(t, res.StatusCode, 200)

	req, _ = http.NewRequest("HEAD", "http://mort/local/fila-test", &buf)
	res = rp.Process(req, obj)

	assert.Equal(t, res.StatusCode, 200)
}

func TestS3GeT(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://mort/local", nil)

	mortConfig := config.Config{}
	err := mortConfig.Load("./benchmark/small.yml")
	assert.Nil(t, err)

	obj, err := object.NewFileObject(req.URL, &mortConfig)
	assert.Nil(t, err)

	rp := NewRequestProcessor(mortConfig.Server, lock.NewMemoryLock(), throttler.NewBucketThrottler(10))
	res := rp.Process(req, obj)

	assert.Equal(t, res.StatusCode, 200)
	assert.Equal(t, res.Headers.Get("content-type"), "application/xml")
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
