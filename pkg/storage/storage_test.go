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
	"github.com/aldor007/mort/pkg/response"
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

	resDel := Delete(obj)

	assert.Equal(t, resDel.StatusCode, 200)
}

func TestHeadS3BucketError(t *testing.T) {
	mortConfig := config.Config{}
	mortConfig.Load("testdata/config2.yml")

	obj, _ := object.NewFileObjectFromPath("/buckets3/file", &mortConfig)

	res := Head(obj)

	assert.Equal(t, res.StatusCode, 500)

	res = Get(obj)

	assert.Equal(t, res.StatusCode, 500)

	res = List(obj, 100, "", "", "")

	assert.Equal(t, res.StatusCode, 500)

	res = Delete(obj)

	assert.Equal(t, res.StatusCode, 500)
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

func TestHeadLocalBucketError(t *testing.T) {
	mortConfig := config.Config{}
	mortConfig.Load("testdata/config2.yml")

	obj, _ := object.NewFileObjectFromPath("/bucketlocal/file", &mortConfig)

	res := Head(obj)

	assert.Equal(t, res.StatusCode, 503)

	res = Get(obj)

	assert.Equal(t, res.StatusCode, 503)

	res = List(obj, 1000, "", "", "")

	assert.Equal(t, res.StatusCode, 503)

	res = Delete(obj)

	assert.Equal(t, res.StatusCode, 503)
}

func TestParseMetadata(t *testing.T) {
	mortConfig := config.Config{}

	mortConfig.Load("testdata/config2.yml")
	obj, _ := object.NewFileObjectFromPath("/buckets3/file", &mortConfig)

	meta := make(map[string]interface{})
	meta["cache-control"] = "max-age=200"
	meta["public"] = "200"

	res := response.NewNoContent(200)
	parseMetadata(obj, meta, res)

	assert.Equal(t, res.Headers.Get("Cache-control"), "max-age=200")
	assert.Equal(t, res.Headers.Get("x-amz-meta-public"), "200")
}

func TestPrepareMetaData(t *testing.T) {
	headers := make(http.Header)
	headers.Add("x-amz-meta-public", "p")
	headers.Add("content-type", "text/html")

	mortConfig := config.Config{}

	mortConfig.Load("testdata/config2.yml")
	obj, _ := object.NewFileObjectFromPath("/buckets3/file", &mortConfig)

	meta := prepareMetadata(obj, headers)

	assert.Equal(t, meta["public"], "p")
	assert.Equal(t, meta["content-type"], "text/html")
}

func TestGetClientAllStorage(t *testing.T) {
	mortConfig := config.Config{}
	mortConfig.Load("testdata/all-storages.yml")
	storages := []string{"local", "http", "s3", "local-meta", "b2", "google", "oracle", "azure"}
	for _, storage := range storages {
		obj, _ := object.NewFileObjectFromPath(fmt.Sprintf("/%s/file", storage), &mortConfig)
		getClient(obj)

	}
}

func TestGetS3(t *testing.T) {
	if os.Getenv("S3_ACCESS_KEY") == "" {
		t.Skip()
	}
	mortConfig := config.Config{}
	mortConfig.Load("testdata/config_s3.yml")
	obj, err := object.NewFileObjectFromPath("/files/sources/2022/Lizbona_2_e38d7c5cac.jpg", &mortConfig)
	assert.NoError(t, err)
	res := Get(obj)
	assert.NoError(t, res.Error())
	assert.Equal(t, 200, res.StatusCode)

	obj, err = object.NewFileObjectFromPath("/images/transform/ZmlsZXMvc291cmNlcy8yMDIyL0xpemJvbmFfMl9lMzhkN2M1Y2FjLmpwZw/photo_Lizbona-2-jpg_big300.jpg", &mortConfig)
	assert.NoError(t, err)
	res = Get(obj)
	assert.NoError(t, res.Error())
	assert.Equal(t, 200, res.StatusCode)
}

func BenchmarkGet(b *testing.B) {
	mortConfig := config.Config{}
	mortConfig.Load("testdata/all-storages.yml")
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
