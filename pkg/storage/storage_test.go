package storage

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/object"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	rootPath := "./testdata"
	os.RemoveAll(filepath.Join(rootPath, "bucket"))
	err := os.Mkdir(filepath.Join(rootPath, "bucket"), 0777)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = ioutil.WriteFile(filepath.Join(rootPath, "bucket", "file"), []byte("3.1"), 0777)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	code := m.Run()

	os.RemoveAll(filepath.Join(rootPath, "bucket"))
	os.Exit(code)
}

func TestGet(t *testing.T) {
	mortConfig := config.Config{}
	mortConfig.Load("testdata/config.yml")

	obj, _ := object.NewFileObjectFromPath("/bucket/file", &mortConfig)

	res := Get(obj)

	assert.Equal(t, res.StatusCode, 200)
}

func TestGetNotFound(t *testing.T) {
	mortConfig := config.Config{}
	mortConfig.Load("testdata/config.yml")

	obj, _ := object.NewFileObjectFromPath("/bucket/file-404", &mortConfig)

	res := Get(obj)

	assert.Equal(t, res.StatusCode, 404)
}

func TestHead(t *testing.T) {
	mortConfig := config.Config{}
	mortConfig.Load("testdata/config.yml")

	obj, _ := object.NewFileObjectFromPath("/bucket/file", &mortConfig)

	res := Head(obj)

	assert.Equal(t, res.StatusCode, 200)
}

func TestHeadNotFound(t *testing.T) {
	mortConfig := config.Config{}
	mortConfig.Load("testdata/config.yml")

	obj, _ := object.NewFileObjectFromPath("/bucket/file-404", &mortConfig)

	res := Head(obj)

	assert.Equal(t, res.StatusCode, 404)
}

func TestDeleteNotFound(t *testing.T) {
	mortConfig := config.Config{}
	mortConfig.Load("testdata/config.yml")

	obj, _ := object.NewFileObjectFromPath("/bucket/file-404", &mortConfig)

	res := Delete(obj)

	assert.Equal(t, res.StatusCode, 200)
}

func TestList(t *testing.T) {
	mortConfig := config.Config{}
	mortConfig.Load("testdata/config.yml")

	obj, _ := object.NewFileObjectFromPath("/bucket/", &mortConfig)

	res := List(obj, 1000, "", "", "")

	assert.Equal(t, res.StatusCode, 200)
	assert.Equal(t, res.Headers.Get("content-type"), "application/xml")
}

func TestSet(t *testing.T) {
	mortConfig := config.Config{}
	mortConfig.Load("testdata/config.yml")

	obj, _ := object.NewFileObjectFromPath("/bucket/file-set", &mortConfig)

	headers := make(http.Header)
	headers["X-Amz-Meta-Header"] = []string{"val"}
	buf := make([]byte, 1000)
	res := Set(obj, headers, int64(len(buf)), ioutil.NopCloser(bytes.NewReader(buf)))

	assert.Equal(t, res.StatusCode, 200)

	resGet := Get(obj)

	assert.Equal(t, resGet.Headers.Get("X-Amz-Meta-Header"), "val")
	assert.Equal(t, resGet.StatusCode, 200)

	resHead := Head(obj)

	assert.Equal(t, resHead.StatusCode, 200)
	assert.Equal(t, resHead.Headers.Get("X-Amz-Meta-Header"), "val")
}

func TestHeadS3BucketError(t *testing.T) {
	mortConfig := config.Config{}
	mortConfig.Load("testdata/config2.yml")

	obj, _ := object.NewFileObjectFromPath("/buckets3/file", &mortConfig)

	res := Head(obj)

	assert.Equal(t, res.StatusCode, 503)

	res = Get(obj)

	assert.Equal(t, res.StatusCode, 503)

	res = List(obj, 100, "", "", "")

	assert.Equal(t, res.StatusCode, 503)

	res = Delete(obj)

	assert.Equal(t, res.StatusCode, 503)
}

func TestHeadHTTPBucketError(t *testing.T) {
	mortConfig := config.Config{}
	mortConfig.Load("testdata/config2.yml")

	obj, _ := object.NewFileObjectFromPath("/buckethttp/file", &mortConfig)

	res := Head(obj)

	assert.Equal(t, res.StatusCode, 500)

	res = Get(obj)

	assert.Equal(t, res.StatusCode, 500)

	res = List(obj, 1000, "", "", "")

	assert.Equal(t, res.StatusCode, 500)

	res = Delete(obj)

	assert.Equal(t, res.StatusCode, 500)
}

func BenchmarkGet(b *testing.B) {
	mortConfig := config.Config{}
	mortConfig.Load("testdata/config.yml")
	obj, _ := object.NewFileObjectFromPath("/bucket/file", &mortConfig)
	for i := 0; i < b.N; i++ {
		Get(obj)
	}
}

func BenchmarkHead(b *testing.B) {
	mortConfig := config.Config{}
	mortConfig.Load("testdata/config.yml")
	obj, _ := object.NewFileObjectFromPath("/bucket/file", &mortConfig)
	for i := 0; i < b.N; i++ {
		Head(obj)
	}
}
