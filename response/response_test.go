package response

import (
	"io/ioutil"
	"bytes"
	"testing"
	"errors"
	"github.com/stretchr/testify/assert"
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
	res.Set("x-header", "1")
	res.SetContentType("text/plain")

	assert.Equal(t, res.StatusCode, 200)
	assert.Equal(t, res.ContentType, "text/plain")
	assert.Equal(t, res.Headers["X-Header"][0], "1")

	buf2, err := res.ReadBody()
	assert.Nil(t, err, "Should not return error when reading body")
	assert.Equal(t, len(buf), len(buf2), "buffors from response should have equal length")
	assert.Equal(t, len(buf), 1000, "buffors from response should have equal length")
}

func TestNewString(t *testing.T) {
	res := NewString(200, "12345")
	res.Set("x-header", "1")
	res.SetContentType("text/plain")

	assert.Equal(t, res.StatusCode, 200)
	assert.Equal(t, res.ContentType, "text/plain")
	assert.Equal(t, res.Headers["X-Header"][0], "1")

	buf2, err := res.ReadBody()
	assert.Nil(t, err, "Should not return error when reading body")
	assert.Equal(t, len(buf2), 5, "buffors from response should have equal length")

}

func TestNewNoContent(t *testing.T) {
	res := NewNoContent(400)
	res.Set("x-header", "1")
	res.SetContentType("text/plain")

	assert.Equal(t, res.StatusCode, 400)
	assert.Equal(t, res.ContentType, "text/plain")
	assert.Equal(t, res.Headers["X-Header"][0], "1")

	buf2, err := res.ReadBody()
	assert.NotNil(t, err, "Should return error when reading body")
	assert.Nil(t, buf2)
}

func TestNewError(t *testing.T) {
	res := NewError(500, errors.New("costam"))
	res.Set("x-header", "1")
	res.SetContentType("text/plain")

	assert.Equal(t, res.StatusCode, 500)
	assert.Equal(t, res.ContentType, "text/plain")
	assert.Equal(t, res.Headers["X-Header"][0], "1")

	buf, err := res.ReadBody()
	assert.NotNil(t, err, "Should return error when reading body")
	assert.Nil(t, buf)

	res.SetDebug(true)
	buf, err = res.ReadBody()
	assert.Nil(t, err)
	assert.NotNil(t, buf, "Should return error when reading body")

}
