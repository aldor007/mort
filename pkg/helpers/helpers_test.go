package helpers

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
	"net/http"
	"sync"
	"testing"

	"github.com/pkg/errors"
)

func TestIsRangeOrCondition(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://url", nil)

	req.Header.Add("range", "0-1")

	assert.True(t, IsRangeOrCondition(req))

	req, _ = http.NewRequest("GET", "http://url", nil)

	req.Header.Add("if-match", "a")

	assert.True(t, IsRangeOrCondition(req))

	req, _ = http.NewRequest("GET", "http://url", nil)

	req.Header.Add("If-Unmodified-Since", "date")

	assert.True(t, IsRangeOrCondition(req))

	req, _ = http.NewRequest("GET", "http://url", nil)
	req.Header.Add("accept-encoding", "gzip")

	assert.False(t, IsRangeOrCondition(req))
}

func TestIsRangeOrCondition_AllHeaders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		header      string
		value       string
		expected    bool
		description string
	}{
		{
			name:        "should detect Range header",
			header:      "Range",
			value:       "bytes=0-1023",
			expected:    true,
			description: "Range header should be detected",
		},
		{
			name:        "should detect If-Range header",
			header:      "If-Range",
			value:       "etag123",
			expected:    true,
			description: "If-Range header should be detected",
		},
		{
			name:        "should detect If-Match header",
			header:      "If-Match",
			value:       "etag456",
			expected:    true,
			description: "If-Match header should be detected",
		},
		{
			name:        "should detect If-None-Match header",
			header:      "If-None-Match",
			value:       "etag789",
			expected:    true,
			description: "If-None-Match header should be detected",
		},
		{
			name:        "should detect If-Modified-Since header",
			header:      "If-Modified-Since",
			value:       "Wed, 21 Oct 2015 07:28:00 GMT",
			expected:    true,
			description: "If-Modified-Since header should be detected",
		},
		{
			name:        "should detect If-Unmodified-Since header",
			header:      "If-Unmodified-Since",
			value:       "Wed, 21 Oct 2015 07:28:00 GMT",
			expected:    true,
			description: "If-Unmodified-Since header should be detected",
		},
		{
			name:        "should not detect regular headers",
			header:      "Accept",
			value:       "application/json",
			expected:    false,
			description: "regular headers should not be detected",
		},
		{
			name:        "should not detect Content-Type",
			header:      "Content-Type",
			value:       "text/html",
			expected:    false,
			description: "Content-Type should not be detected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req, _ := http.NewRequest("GET", "http://example.com", nil)
			req.Header.Add(tt.header, tt.value)

			result := IsRangeOrCondition(req)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestIsRangeOrCondition_MultipleHeaders(t *testing.T) {
	t.Parallel()

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	req.Header.Add("Range", "bytes=0-1023")
	req.Header.Add("If-Match", "etag123")

	result := IsRangeOrCondition(req)
	assert.True(t, result, "should detect when multiple range/conditional headers are present")
}

func TestIsRangeOrCondition_NoHeaders(t *testing.T) {
	t.Parallel()

	req, _ := http.NewRequest("GET", "http://example.com", nil)

	result := IsRangeOrCondition(req)
	assert.False(t, result, "should return false when no range/conditional headers")
}

func TestFetchUrlObject(t *testing.T) {
	defer gock.Off()

	gock.New("http://image.om").
		Get("/bar.jpg").
		Reply(200).
		BodyString("foo foo")

	gock.InterceptClient(client)
	buf, err := FetchObject("http://image.om/bar.jpg")

	assert.Nil(t, err)
	assert.Equal(t, string(buf), "foo foo")
}

func TestFetchUrlObjectErr(t *testing.T) {
	defer gock.Off()

	gock.New("http://image.om").
		Get("/bar.jpg").
		ReplyError(errors.New("error"))

	gock.InterceptClient(client)
	_, err := FetchObject("http://image.om/bar.jpg")

	assert.NotNil(t, err)
}

func TestFetchObjectErr(t *testing.T) {
	_, err := FetchObject("bar.jpg")

	assert.NotNil(t, err)
}

func TestFetchObject(t *testing.T) {
	_, err := FetchObject("./helpers.go")

	assert.Nil(t, err)
}

func TestFetchObject_HTTPSSuccess(t *testing.T) {
	defer gock.Off()

	gock.New("https://secure.example.com").
		Get("/image.png").
		Reply(200).
		BodyString("secure content")

	gock.InterceptClient(client)
	buf, err := FetchObject("https://secure.example.com/image.png")

	assert.Nil(t, err, "should fetch HTTPS URL successfully")
	assert.Equal(t, "secure content", string(buf))
}

func TestFetchObject_HTTPWithPath(t *testing.T) {
	defer gock.Off()

	gock.New("http://example.com").
		Get("/path/to/resource.jpg").
		Reply(200).
		BodyString("path content")

	gock.InterceptClient(client)
	buf, err := FetchObject("http://example.com/path/to/resource.jpg")

	assert.Nil(t, err, "should fetch URL with path")
	assert.Equal(t, "path content", string(buf))
}

func TestFetchObject_LocalFileWithPath(t *testing.T) {
	t.Parallel()

	buf, err := FetchObject("./helpers.go")

	assert.Nil(t, err, "should read local file")
	assert.NotEmpty(t, buf, "should have file content")
	assert.Contains(t, string(buf), "package helpers", "should contain package declaration")
}

func TestFetchObject_LocalFileNotFound(t *testing.T) {
	t.Parallel()

	_, err := FetchObject("./nonexistent-file.txt")

	assert.NotNil(t, err, "should return error for missing file")
}

func TestFetchObject_InvalidURL(t *testing.T) {
	defer gock.Off()

	gock.New("http://invalid-url-that-does-not-exist-12345.com").
		Get("/test").
		ReplyError(errors.New("invalid URL"))

	gock.InterceptClient(client)
	_, err := FetchObject("http://invalid-url-that-does-not-exist-12345.com/test")

	assert.NotNil(t, err, "should return error for invalid URL")
}

func TestFetchObject_LargeResponse(t *testing.T) {
	defer gock.Off()

	// Create large content (1KB)
	largeContent := ""
	for i := 0; i < 1024; i++ {
		largeContent += string(byte('A' + (i % 26)))
	}

	gock.New("http://example.com").
		Get("/large.bin").
		Reply(200).
		BodyString(largeContent)

	gock.InterceptClient(client)
	buf, err := FetchObject("http://example.com/large.bin")

	assert.Nil(t, err, "should fetch large response")
	assert.Equal(t, len(largeContent), len(buf), "should receive all content")
}

func TestFetchObject_EmptyResponse(t *testing.T) {
	defer gock.Off()

	gock.New("http://example.com").
		Get("/empty").
		Reply(200).
		BodyString("")

	gock.InterceptClient(client)
	buf, err := FetchObject("http://example.com/empty")

	assert.Nil(t, err, "should handle empty response")
	assert.Equal(t, 0, len(buf), "should have empty content")
}

func TestFetchObject_HTTPStatusCodes(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		description string
	}{
		{
			name:        "should handle 404 Not Found",
			statusCode:  404,
			description: "404 response should be handled",
		},
		{
			name:        "should handle 500 Internal Server Error",
			statusCode:  500,
			description: "500 response should be handled",
		},
		{
			name:        "should handle 301 redirect",
			statusCode:  301,
			description: "301 response should be handled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer gock.Off()

			gock.New("http://example.com").
				Get("/test").
				Reply(tt.statusCode).
				BodyString("response body")

			gock.InterceptClient(client)
			buf, err := FetchObject("http://example.com/test")

			// FetchObject doesn't check status codes, it just reads the body
			assert.Nil(t, err, tt.description)
			assert.Equal(t, "response body", string(buf))
		})
	}
}

func TestFetchObject_ConcurrentFetches(t *testing.T) {
	// Note: Cannot use t.Parallel() with gock as it uses global state
	defer gock.Off()

	// Set up mock for multiple requests
	for i := 0; i < 10; i++ {
		gock.New("http://example.com").
			Get("/concurrent").
			Reply(200).
			BodyString("concurrent content")
	}

	gock.InterceptClient(client)

	var wg sync.WaitGroup
	errors := make([]error, 10)
	results := make([][]byte, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			buf, err := FetchObject("http://example.com/concurrent")
			results[index] = buf
			errors[index] = err
		}(i)
	}

	wg.Wait()

	// Verify all requests succeeded
	for i := 0; i < 10; i++ {
		assert.Nil(t, errors[i], "concurrent request %d should succeed", i)
		assert.Equal(t, "concurrent content", string(results[i]), "should have correct content")
	}
}

func TestFetchObject_HTTPvsLocal(t *testing.T) {
	tests := []struct {
		name        string
		uri         string
		shouldError bool
		description string
	}{
		{
			name:        "should distinguish HTTP URL",
			uri:         "http://example.com/file.jpg",
			shouldError: false, // Will be mocked
			description: "HTTP URL should use HTTP client",
		},
		{
			name:        "should distinguish HTTPS URL",
			uri:         "https://example.com/file.jpg",
			shouldError: false, // Will be mocked
			description: "HTTPS URL should use HTTP client",
		},
		{
			name:        "should treat non-HTTP as local file",
			uri:         "./helpers.go",
			shouldError: false,
			description: "local path should use file system",
		},
		{
			name:        "should treat absolute path as local file",
			uri:         "/nonexistent/file.txt",
			shouldError: true,
			description: "absolute path should use file system",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.uri == "http://example.com/file.jpg" || tt.uri == "https://example.com/file.jpg" {
				defer gock.Off()
				if tt.uri[0:5] == "https" {
					gock.New("https://example.com").Get("/file.jpg").Reply(200).BodyString("mocked")
				} else {
					gock.New("http://example.com").Get("/file.jpg").Reply(200).BodyString("mocked")
				}
				gock.InterceptClient(client)
			}

			_, err := FetchObject(tt.uri)

			if tt.shouldError {
				assert.NotNil(t, err, tt.description)
			} else {
				assert.Nil(t, err, tt.description)
			}
		})
	}
}

func TestFetchObject_ClientReuse(t *testing.T) {
	defer gock.Off()

	// Multiple requests should reuse the same client
	gock.New("http://example.com").
		Get("/test1").
		Reply(200).
		BodyString("test1")

	gock.New("http://example.com").
		Get("/test2").
		Reply(200).
		BodyString("test2")

	gock.InterceptClient(client)

	buf1, err1 := FetchObject("http://example.com/test1")
	assert.Nil(t, err1)
	assert.Equal(t, "test1", string(buf1))

	buf2, err2 := FetchObject("http://example.com/test2")
	assert.Nil(t, err2)
	assert.Equal(t, "test2", string(buf2))
}
