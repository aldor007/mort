package processor

import (
	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/lock"
	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/throttler"
	//"github.com/aldor007/mort/pkg/monitoring"
	//"go.uber.org/zap"
	"github.com/stretchr/testify/assert"
	"net/http"
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
	assert.Equal(t, res.Headers.Get("x-amz-meta-public-width"), "150")
	assert.Equal(t, res.Headers.Get("ETag"), "W/\"7eaa484e8c841e7e\"")
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

	//logger, _ := zap.NewProduction()
	////logger, _ := zap.NewDevelopment()
	//zap.ReplaceGlobals(logger)
	//log.RegisterLogger(logger)
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
					b.Fatalf("Invalid response sc %s test name %s", res.StatusCode, bm.name)
				}
			}
		})
	}

}
