package response

import (
	"bytes"
	"compress/gzip"
	"errors"
	"github.com/aldor007/mort/pkg/helpers"
	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/transforms"
	"github.com/stretchr/testify/assert"
	"github.com/vmihailenco/msgpack"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResponse_Copy(t *testing.T) {
	buf := make([]byte, 1000)
	res := NewBuf(200, buf)
	resCpy, err := res.Copy()
	assert.Nil(t, err, "Should not return error when copying")

	assert.Equal(t, res.StatusCode, resCpy.StatusCode, "status code should be equal")
	assert.Equal(t, res.ContentLength, resCpy.ContentLength, "content type code should be equal")

	buf1, err := res.Body()
	assert.Nil(t, err, "Should not return error when reading body")
	buf2, err := resCpy.Body()
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

	buf2, err := res.Body()
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

	buf2, err := res.Body()
	assert.Nil(t, err, "Should not return error when reading body")
	assert.Equal(t, len(buf2), 5, "buffors from response should have equal length")

}

func TestNewNoContent(t *testing.T) {
	res := NewNoContent(400)
	res.Set("x-header", "1")
	res.SetContentType("text/plain")

	assert.Equal(t, res.StatusCode, 400)
	assert.Equal(t, res.Headers.Get(HeaderContentType), "text/plain")
	assert.Equal(t, res.Headers["X-Header"][0], "1")

	buf2, err := res.Body()
	assert.NotNil(t, err, "Should return error when reading body")
	assert.Nil(t, buf2)
}

func TestNewError(t *testing.T) {
	err := errors.New("costam")
	res := NewError(500, err)
	res.Headers.Set("x-header", "1")
	res.SetContentType("text/plain")

	assert.Equal(t, res.StatusCode, 500)
	assert.Equal(t, res.Headers.Get(HeaderContentType), "text/plain")
	assert.Equal(t, res.Headers["X-Header"][0], "1")
	assert.True(t, res.HasError())
	assert.Equal(t, res.Error(), err)

	buf, err := res.Body()
	assert.NotNil(t, err, "Should return error when reading body")
	assert.Nil(t, buf)

	p := &object.FileObject{}
	res.SetDebug(&object.FileObject{Debug: true, Parent: p, Transforms: transforms.Transforms{NotEmpty: true}})
	buf, err = res.Body()
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
	assert.True(t, res.IsBuffered())
	assert.Equal(t, result.StatusCode, 200)
	assert.Equal(t, result.Header.Get("X-Header"), "1")
	assert.Equal(t, result.Header.Get("Content-Type"), "text/html")
}

func TestResponse_Send_and_Copy(t *testing.T) {
	buf := make([]byte, 1000)
	res := New(200, ioutil.NopCloser(bytes.NewReader(buf)))
	res.Headers.Set("X-Header", "1")
	res.SetContentType("text/html")

	resCpy, err := res.Copy()
	assert.Nil(t, err)
	recorder := httptest.NewRecorder()
	res.Send(recorder)

	result := recorder.Result()
	assert.Equal(t, result.StatusCode, 200)
	assert.Equal(t, result.Header.Get("X-Header"), "1")
	assert.Equal(t, result.Header.Get("Content-Type"), "text/html")

	assert.Equal(t, resCpy.StatusCode, 200)
	assert.Equal(t, resCpy.Headers.Get("X-Header"), "1")
	assert.Equal(t, resCpy.Headers.Get("Content-Type"), "text/html")
	body, err := resCpy.Body()

	assert.Nil(t, err, "Shouldn't return error when reading body")
	assert.Equal(t, len(body), 1000)
}

func TestResponse_SendContentNotRangeOrCondition(t *testing.T) {
	buf := make([]byte, 1000)
	res := New(200, ioutil.NopCloser(bytes.NewReader(buf)))
	res.Headers.Set("X-Header", "1")
	res.SetContentType("text/html")

	req, _ := http.NewRequest("GET", "/bucket/local.jpg", nil)
	recorder := httptest.NewRecorder()
	res.SendContent(req, recorder)

	result := recorder.Result()
	assert.Equal(t, result.StatusCode, 200)
	assert.Equal(t, result.Header.Get("X-Header"), "1")
	body, _ := ioutil.ReadAll(result.Body)
	assert.Equal(t, len(body), 1000)
}

func TestResponse_BodyTransformer(t *testing.T) {
	buf := make([]byte, 1000)
	res := New(200, ioutil.NopCloser(bytes.NewReader(buf)))
	res.Headers.Set("X-Header", "1")
	res.SetContentType("text/html")

	var writerCalled bool

	res.BodyTransformer(func(writer io.Writer) io.WriteCloser {
		writerCalled = true
		gzipW := gzip.NewWriter(writer)
		return gzipW
	})

	recorder := httptest.NewRecorder()
	res.Send(recorder)

	assert.True(t, writerCalled)
}

func TestResponse_IsImage(t *testing.T) {
	res := NewNoContent(299)

	res.Headers.Set("Content-type", "image/jpg")

	assert.True(t, res.IsImage())
}

func TestIsRangeOrCond(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://url", nil)
	req.Header.Set("Range", "1-3")
	assert.True(t, helpers.IsRangeOrCondition(req))

	req, _ = http.NewRequest("GET", "http://url", nil)
	req.Header.Set("If-match", "1-3")
	assert.True(t, helpers.IsRangeOrCondition(req))

	req, _ = http.NewRequest("GET", "http://url", nil)
	req.Header.Set("If-unmodified-since", "1-3")
	assert.True(t, helpers.IsRangeOrCondition(req))

	req, _ = http.NewRequest("GET", "http://url", nil)
	assert.False(t, helpers.IsRangeOrCondition(req))
}

func TestResponse_IsCachable(t *testing.T) {
	res := NewNoContent(200)

	assert.False(t, res.IsCacheable())

	res.Headers.Set("cache-control", "public, max-age=600")

	assert.True(t, res.IsCacheable())
	assert.Equal(t, res.GetTTL(), 600)
}

func TestResponse_DecodeMsgpack(t *testing.T) {
	res := NewString(400, "testuje")
	res.Headers.Set("etag", "md5")
	res.Headers.Set("cache-control", "private")
	res.Body()
	buf, err := msgpack.Marshal(res)

	assert.Nil(t, err)

	var resMsg Response
	err = msgpack.Unmarshal(buf, &resMsg)

	assert.Nil(t, err)

	b, err := resMsg.Body()
	assert.Nil(t, err)

	assert.Equal(t, resMsg.StatusCode, 400)
	assert.Equal(t, resMsg.Headers.Get("etag"), "md5")
	assert.Equal(t, resMsg.GetTTL(), -1)
	assert.Equal(t, string(b), "testuje")

}

func BenchmarkNewBuf(b *testing.B) {
	buf := make([]byte, 1000)
	for i := 0; i < b.N; i++ {
		res := NewBuf(200, buf)
		res.Headers.Set("X-Header", "1")
		res.SetContentType("text/html")
	}
}

func BenchmarkNewCopy(b *testing.B) {
	buf := make([]byte, 1024*1024*4)
	for i := 0; i < b.N; i++ {
		s := ioutil.NopCloser(bytes.NewReader(buf))
		res := New(200, s)
		res.Headers.Set("X-Header", "1")
		res.SetContentType("text/html")
		resCpy, _ := res.Copy()

		body, err := resCpy.Body()
		if err != nil {
			b.Fatalf("Errors %s", err)
		}

		if len(body) != len(buf) {
			b.Fatalf("Inavlid body len %d != %d %d", len(body), len(buf), i)
		}
	}
}

func BenchmarkNewCopyWithStream(b *testing.B) {
	buf := make([]byte, 1024*1024*4)
	wBuf := make([]byte, 0, 1024*1024*1)
	for i := 0; i < b.N; i++ {
		s := ioutil.NopCloser(bytes.NewReader(buf))
		w := bytes.NewBuffer(wBuf)
		res := New(200, s)
		res.Headers.Set("X-Header", "1")
		res.SetContentType("text/html")
		resCpy, _ := res.CopyWithStream()
		go func() {
			io.Copy(w, res.Stream())
			res.Close()
		}()

		body, err := ioutil.ReadAll(resCpy.Stream())
		if err != nil {
			b.Fatalf("Errors %s", err)
		}

		if len(body) != len(buf) {
			b.Fatalf("Inavlid body len %d != %d %d", len(body), len(buf), i)
		}
	}
}
