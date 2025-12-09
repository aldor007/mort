package storage

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"
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

// TestStorageCacheConcurrency tests that concurrent access to storage cache
// doesn't cause race conditions or lock contention issues
func TestStorageCacheConcurrency(t *testing.T) {
	// Note: Cannot use t.Parallel() because this test resets the global storageCache

	mortConfig := config.Config{}
	err := mortConfig.Load("testdata/config.yml")
	assert.Nil(t, err)

	// Clear storage cache to ensure fresh state
	storageCache = sync.Map{}

	// Create multiple goroutines that all try to get storage clients concurrently
	// This should trigger the race condition that was fixed with sync.Once
	var wg sync.WaitGroup
	numGoroutines := 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			obj, err := object.NewFileObjectFromPath("/bucket/file", &mortConfig)
			assert.Nil(t, err)

			// Call getClient multiple times
			for j := 0; j < 10; j++ {
				client, err := getClient(obj)
				assert.Nil(t, err)
				assert.NotNil(t, client.client)
				assert.NotNil(t, client.container)
			}
		}()
	}

	wg.Wait()

	// Verify that only one storage client was created (stored in cache)
	// Count entries in sync.Map
	count := 0
	storageCache.Range(func(key, value interface{}) bool {
		count++
		entry := value.(*storageClientEntry)
		assert.Nil(t, entry.err)
		assert.NotNil(t, entry.client.client)
		return true
	})

	// Should have exactly one entry for the storage config hash
	assert.Equal(t, 1, count, "Should have exactly one cached storage client")
}

func TestSetWithInvalidPath(t *testing.T) {
	t.Parallel()

	mortConfig := config.Config{}
	mortConfig.Load("testdata/config.yml")

	obj, _ := object.NewFileObjectFromPath("/bucket/../invalid", &mortConfig)

	headers := make(http.Header)
	buf := make([]byte, 10)
	res := Set(obj, headers, int64(len(buf)), ioutil.NopCloser(bytes.NewReader(buf)))

	// Should still succeed with local storage
	assert.Equal(t, 200, res.StatusCode)
}

func TestSetWithSmallContent(t *testing.T) {
	t.Parallel()

	mortConfig := config.Config{}
	mortConfig.Load("testdata/config.yml")

	obj, _ := object.NewFileObjectFromPath("/bucket/small-file", &mortConfig)

	headers := make(http.Header)
	buf := []byte("x")
	res := Set(obj, headers, int64(len(buf)), ioutil.NopCloser(bytes.NewReader(buf)))

	assert.Equal(t, 200, res.StatusCode)

	// Verify file exists
	resGet := Get(obj)
	assert.Equal(t, 200, resGet.StatusCode)
	body, _ := resGet.Body()
	assert.Equal(t, 1, len(body))

	// Cleanup
	Delete(obj)
}

func TestGetWithRange(t *testing.T) {
	t.Parallel()

	mortConfig := config.Config{}
	mortConfig.Load("testdata/config.yml")

	obj, _ := object.NewFileObjectFromPath("/bucket/file", &mortConfig)
	obj.Range = "bytes=0-1"

	res := Get(obj)
	assert.Equal(t, 200, res.StatusCode)
}

func TestSetAndGetLargeFile(t *testing.T) {
	t.Parallel()

	mortConfig := config.Config{}
	mortConfig.Load("testdata/config.yml")

	obj, _ := object.NewFileObjectFromPath("/bucket/large-file", &mortConfig)

	headers := make(http.Header)
	headers.Add("Content-Type", "application/octet-stream")
	// Create a 100KB file
	buf := make([]byte, 100*1024)
	for i := range buf {
		buf[i] = byte(i % 256)
	}

	res := Set(obj, headers, int64(len(buf)), ioutil.NopCloser(bytes.NewReader(buf)))
	assert.Equal(t, 200, res.StatusCode)

	// Verify file was stored correctly
	resGet := Get(obj)
	assert.Equal(t, 200, resGet.StatusCode)
	body, _ := resGet.Body()
	assert.Equal(t, len(buf), len(body))

	// Verify content matches
	assert.Equal(t, buf[0], body[0])
	assert.Equal(t, buf[len(buf)-1], body[len(body)-1])

	// Cleanup
	Delete(obj)
}

func TestListWithMaxKeys(t *testing.T) {
	// Note: Cannot use t.Parallel() because List reads filesystem state
	// that may be modified by other parallel tests

	mortConfig := config.Config{}
	mortConfig.Load("testdata/config.yml")

	// List with max keys limit
	obj, _ := object.NewFileObjectFromPath("/bucket/", &mortConfig)
	res := List(obj, 10, "", "", "")

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "application/xml", res.Headers.Get("content-type"))
}

func TestGetKey(t *testing.T) {
	t.Parallel()

	mortConfig := config.Config{}
	mortConfig.Load("testdata/config.yml")

	obj, _ := object.NewFileObjectFromPath("/bucket/test/file.jpg", &mortConfig)
	key := getKey(obj)

	// For local storage, key includes the path
	assert.NotEmpty(t, key)
	assert.Contains(t, key, "test/file.jpg")
}

func TestPrepareMetadataWithContentType(t *testing.T) {
	t.Parallel()

	headers := make(http.Header)
	headers.Add("Content-Type", "image/jpeg")
	headers.Add("X-Amz-Meta-Custom", "value")
	headers.Add("Etag", "abc123")

	mortConfig := config.Config{}
	mortConfig.Load("testdata/config.yml")
	obj, _ := object.NewFileObjectFromPath("/bucket/file.jpg", &mortConfig)

	meta := prepareMetadata(obj, headers)

	// prepareMetadata only stores content-type, etag, and x-amz-meta headers for local storage
	assert.Equal(t, "image/jpeg", meta["content-type"])
	assert.Equal(t, "value", meta["x-amz-meta-custom"])
	assert.Equal(t, "abc123", meta["etag"])
}

func TestParseMetadataWithCacheControl(t *testing.T) {
	t.Parallel()

	mortConfig := config.Config{}
	mortConfig.Load("testdata/config.yml")
	obj, _ := object.NewFileObjectFromPath("/bucket/file", &mortConfig)

	meta := make(map[string]interface{})
	meta["cache-control"] = "public, max-age=3600"
	meta["x-custom-header"] = "value"

	res := response.NewNoContent(200)
	parseMetadata(obj, meta, res)

	// parseMetadata handles cache-control and x- prefixed headers
	assert.Equal(t, "public, max-age=3600", res.Headers.Get("Cache-Control"))
	assert.Equal(t, "value", res.Headers.Get("X-Custom-Header"))
}

func TestSetWithMultipleMetadataHeaders(t *testing.T) {
	t.Parallel()

	mortConfig := config.Config{}
	mortConfig.Load("testdata/config.yml")

	obj, _ := object.NewFileObjectFromPath("/bucket/metadata-test", &mortConfig)

	headers := make(http.Header)
	headers.Add("X-Amz-Meta-Author", "TestUser")
	headers.Add("X-Amz-Meta-Version", "1.0")
	headers.Add("Content-Type", "application/json")
	headers.Add("Cache-Control", "no-cache")

	buf := []byte(`{"test": "data"}`)
	res := Set(obj, headers, int64(len(buf)), ioutil.NopCloser(bytes.NewReader(buf)))
	assert.Equal(t, 200, res.StatusCode)

	// Verify metadata was stored
	resHead := Head(obj)
	assert.Equal(t, 200, resHead.StatusCode)
	assert.Equal(t, "TestUser", resHead.Headers.Get("X-Amz-Meta-Author"))
	assert.Equal(t, "1.0", resHead.Headers.Get("X-Amz-Meta-Version"))
	assert.Equal(t, "application/json", resHead.Headers.Get("Content-Type"))

	// Cleanup
	Delete(obj)
}

func TestDeleteMultipleTimes(t *testing.T) {
	t.Parallel()

	mortConfig := config.Config{}
	mortConfig.Load("testdata/config.yml")

	obj, _ := object.NewFileObjectFromPath("/bucket/delete-test", &mortConfig)

	// Create file
	headers := make(http.Header)
	buf := []byte("test")
	Set(obj, headers, int64(len(buf)), ioutil.NopCloser(bytes.NewReader(buf)))

	// Delete first time
	res := Delete(obj)
	assert.Equal(t, 200, res.StatusCode)

	// Delete second time (file already deleted)
	res = Delete(obj)
	assert.Equal(t, 200, res.StatusCode, "Deleting non-existent file should return 200")
}
