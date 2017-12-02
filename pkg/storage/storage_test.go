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

func TestHead(t *testing.T) {
	mortConfig := config.Config{}
	mortConfig.Load("testdata/config.yml")

	obj, _ := object.NewFileObjectFromPath("/bucket/file", &mortConfig)

	res := Head(obj)

	assert.Equal(t, res.StatusCode, 200)
}

func TestSet(t *testing.T) {
	mortConfig := config.Config{}
	mortConfig.Load("testdata/config.yml")

	obj, _ := object.NewFileObjectFromPath("/bucket/file-set", &mortConfig)

	headers := make(http.Header)
	headers["X-Header"] = []string{"val"}
	buf := make([]byte, 1000)
	res := Set(obj, headers, int64(len(buf)), ioutil.NopCloser(bytes.NewReader(buf)))

	assert.Equal(t, res.StatusCode, 200)

	resGet := Get(obj)

	assert.Equal(t, resGet.Headers.Get("X-Header"), "val")
	assert.Equal(t, resGet.StatusCode, 200)

	resHead := Head(obj)

	assert.Equal(t, resHead.StatusCode, 200)
	assert.Equal(t, resHead.Headers.Get("X-Header"), "val")
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
