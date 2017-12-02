package processor

import (
	"bytes"
	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/lock"
	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/throttler"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestNewRequestProcessor(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://mort/local/small.jpg-small", nil)

	mortConfig := config.Config{}
	err := mortConfig.Load("./benchmark/small.yml")
	assert.Nil(t, err)

	obj, err := object.NewFileObject(req.URL, &mortConfig)
	assert.Nil(t, err)

	rp := NewRequestProcessor(mortConfig.Server, lock.NewMemoryLock(), throttler.NewBucketThrottler(10))
	res := rp.Process(req, obj)

	assert.Equal(t, res.StatusCode, 200)
	assert.Equal(t, res.Headers.Get("x-amz-meta-public-width"), "105")
	assert.Equal(t, res.Headers.Get("ETag"), "ff6783002d849307")
}

func BenchmarkNewRequestProcessorMemoryLock(b *testing.B) {
	benchmarks := []struct {
		name       string
		url        string
		filePath   string
		configPath string
	}{
		{"Process small image, small result", "http://mort/local/small.jpg-small", "./benchmark/local/small.jpg", "./benchmark/small.yml"},
		{"Process large image, small result", "http://mort/local/large.jpeg-small", "./benchmark/local/large.jpeg", "./benchmark/small.yml"},
	}

	for _, bm := range benchmarks {
		data, err := ioutil.ReadFile(bm.filePath)
		if err != nil {
			panic(err)
		}
		req, _ := http.NewRequest("GET", bm.url, ioutil.NopCloser(bytes.NewReader(data)))

		mortConfig := config.Config{}
		err = mortConfig.Load(bm.configPath)
		if err != nil {
			panic(err)
		}

		obj, _ := object.NewFileObject(req.URL, &mortConfig)
		rp := NewRequestProcessor(mortConfig.Server, lock.NewMemoryLock(), throttler.NewBucketThrottler(10))
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

func BenchmarkNewRequestProcessorNopLock(b *testing.B) {
	benchmarks := []struct {
		name       string
		url        string
		filePath   string
		configPath string
	}{
		{"Process small image, small result", "http://mort/pkg/local/small.jpg-small", "./tests/benchmark/local/small.jpg", "./tests/benchmark/small.yml"},
		{"Process large image, small result", "http://mort/pkg/local/large.jpeg-small", "./tests/benchmark/local/large.jpeg", "./tests/benchmark/small.yml"},
	}

	for _, bm := range benchmarks {
		data, err := ioutil.ReadFile(bm.filePath)
		if err != nil {
			panic(err)
		}
		req, _ := http.NewRequest("GET", bm.url, ioutil.NopCloser(bytes.NewReader(data)))

		mortConfig := config.Config{}
		err = mortConfig.Load(bm.configPath)
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
