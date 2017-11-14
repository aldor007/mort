package response

import (
	"io/ioutil"
	"bytes"
	"testing"
	"errors"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
)

func TestResponse_Copy(t *testing.T) {
	buf := make([]byte, 1000)
	res := NewBuf(200, buf)
	resCpy, err := res.Copy()
	assert.Nil(t, err, "Should not return error when copying")

	assert.Equal(t, res.StatusCode, resCpy.StatusCode, "status code should be equal")
	assert.Equal(t, res.ContentLength, resCpy.ContentLength, "content type code should be equal")

	buf1, err := res.ReadBody()
	assert.Nil(t, err, "Should not return error when reading body")
	buf2, err := resCpy.ReadBody()
	assert.Nil(t, err, "Should not return error when reading body")
	assert.Equal(t, len(buf1), len(buf2), "buffors from response should have equal length")
	assert.Equal(t, len(buf1), 1000, "buffors from response should have equal length")
}

func TestNew(t *testing.T) {
	buf := make([]byte, 1000)
	reader := ioutil.NopCloser(bytes.NewReader(buf))

	res := New(200, reader)
	res.Headers.Set("x-header", "1")
	res.SetContentType("text/plain")

	assert.Equal(t, res.StatusCode, 200)
	assert.Equal(t, res.Headers.Get(HeaderContentType), "text/plain")
	assert.Equal(t, res.Headers["X-Header"][0], "1")

	buf2, err := res.ReadBody()
	assert.Nil(t, err, "Should not return error when reading body")
	assert.Equal(t, len(buf), len(buf2), "buffors from response should have equal length")
	assert.Equal(t, len(buf), 1000, "buffors from response should have equal length")
}

func TestNewString(t *testing.T) {
	res := NewString(200, "12345")
	res.Headers.Set("x-header", "1")
	res.SetContentType("text/plain")

	assert.Equal(t, res.StatusCode, 200)
	assert.Equal(t, res.Headers.Get(HeaderContentType), "text/plain")
	assert.Equal(t, res.Headers["X-Header"][0], "1")

	buf2, err := res.ReadBody()
	assert.Nil(t, err, "Should not return error when reading body")
	assert.Equal(t, len(buf2), 5, "buffors from response should have equal length")

}

func TestNewNoContent(t *testing.T) {
	res := NewNoContent(400)
	res.Headers.Set("x-header", "1")
	res.SetContentType("text/plain")

	assert.Equal(t, res.StatusCode, 400)
	assert.Equal(t, res.Headers.Get(HeaderContentType), "text/plain")
	assert.Equal(t, res.Headers["X-Header"][0], "1")

	buf2, err := res.ReadBody()
	assert.NotNil(t, err, "Should return error when reading body")
	assert.Nil(t, buf2)
}

func TestNewError(t *testing.T) {
	res := NewError(500, errors.New("costam"))
	res.Headers.Set("x-header", "1")
	res.SetContentType("text/plain")

	assert.Equal(t, res.StatusCode, 500)
	assert.Equal(t, res.Headers.Get(HeaderContentType), "text/plain")
	assert.Equal(t, res.Headers["X-Header"][0], "1")

	buf, err := res.ReadBody()
	assert.NotNil(t, err, "Should return error when reading body")
	assert.Nil(t, buf)

	res.SetDebug(true)
	buf, err = res.ReadBody()
	assert.Nil(t, err)
	assert.NotNil(t, buf, "Should return error when reading body")

}

func TestResponse_CopyHeadersFrom(t *testing.T) {
	buf := make([]byte, 1000)
	src := NewBuf(200, buf)
	src.Headers.Set("X-Header", "1")
	src.SetContentType("text/html")

	res := NewBuf(200, buf)
	res.CopyHeadersFrom(src)

	assert.Equal(t, res.Headers["X-Header"][0], "1")
	assert.Equal(t, res.Headers.Get(HeaderContentType), "text/html")
}

func TestResponse_Send(t *testing.T) {
	buf := make([]byte, 1000)
	res := NewBuf(200, buf)
	res.Headers.Set("X-Header", "1")
	res.SetContentType("text/html")

	recorder := httptest.NewRecorder()
	res.Send(recorder)

	result := recorder.Result()
	assert.Equal(t, result.StatusCode, 200)
	assert.Equal(t, result.Header.Get("X-Header"), "1")
	assert.Equal(t, result.Header.Get("Content-Type"), "text/html")
}

func BenchmarkNewBuf(b *testing.B) {
	buf := make([]byte, 1000)
	for i := 0; i < b.N; i++  {
		res := NewBuf(200, buf)
		res.Headers.Set("X-Header", "1")
		res.SetContentType("text/html")
	}
}