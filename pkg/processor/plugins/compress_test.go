package plugins

import (
	"bytes"
	"compress/gzip"
	"github.com/aldor007/mort/pkg/response"
	"github.com/stretchr/testify/assert"
	brEnc "github.com/google/brotli/go/cbrotli"
	"gopkg.in/yaml.v2"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCompressNoAccept(t *testing.T) {
	c := CompressPlugin{}
	configStr := `
    gzip:
       level: 5
`

	var config interface{}
	err := yaml.Unmarshal([]byte(configStr), &config)
	if err != nil {
		panic(err)
	}

	c.configure(config)
	req, _ := http.NewRequest("GET", "http://mort/local/small.jpg-m", nil)
	res := response.NewNoContent(200)

	c.postProcess(nil, req, res)

	assert.Equal(t, len(res.Headers), 0)
}

func TestCompressNoContent(t *testing.T) {
	c := CompressPlugin{}
	configStr := `
    gzip:
       level: 5
`
	var config interface{}
	yaml.Unmarshal([]byte(configStr), &config)

	c.configure(config)
	req, _ := http.NewRequest("GET", "http://mort/local/small.jpg-m", nil)
	req.Header.Add("Accept-Encoding", "gzip")
	res := response.NewNoContent(200)
	res.Headers.Add("Content-Type", "text/html")

	c.postProcess(nil, req, res)

	assert.Equal(t, len(res.Headers), 1)
}

func TestCompressTooSmallContent(t *testing.T) {
	c := CompressPlugin{}
	configStr := `
    gzip:
       level: 5
`
	var config interface{}
	yaml.Unmarshal([]byte(configStr), &config)

	c.configure(config)
	req, _ := http.NewRequest("GET", "http://mort/local/small.jpg-m", nil)
	req.Header.Add("Accept-Encoding", "gzip")
	res := response.NewBuf(200, make([]byte, 10))
	res.Headers.Add("Content-Type", "text/html")

	c.postProcess(nil, req, res)

	assert.Equal(t, len(res.Headers), 1)
}

func TestCompressGzipImage(t *testing.T) {
	c := CompressPlugin{}
	configStr := `
    gzip:
       level: 5
`
	var config interface{}
	yaml.Unmarshal([]byte(configStr), &config)

	c.configure(config)
	req, _ := http.NewRequest("GET", "http://mort/local/small.jpg-m", nil)
	req.Header.Add("Accept-Encoding", "gzip")
	res := response.NewBuf(200, make([]byte, 13000))
	res.Headers.Add("Content-Type", "image/jpg")

	c.postProcess(nil, req, res)

	assert.Equal(t, len(res.Headers), 1)
}

func TestCompressGzip(t *testing.T) {
	c := CompressPlugin{}
	configStr := `
    gzip:
       level: 5
`
	var config interface{}
	yaml.Unmarshal([]byte(configStr), &config)

	c.configure(config)
	req, _ := http.NewRequest("GET", "http://mort/local/small.jpg-m", nil)
	req.Header.Add("Accept-Encoding", "gzip")
	body := make([]byte, 1200)
	body[33] = 'a'
	body[324] = 'c'
	res := response.NewBuf(200, body)
	res.Headers.Add("Content-Type", "text/html")

	c.postProcess(nil, req, res)

	assert.Equal(t, len(res.Headers), 3)
	assert.Equal(t, res.Headers.Get("Content-Encoding"), "gzip")
	assert.Equal(t, res.Headers.Get("Vary"), "Accept-Encoding")

	recorder := httptest.NewRecorder()
	res.Send(recorder)

	var buf bytes.Buffer
	gzipW, _ := gzip.NewWriterLevel(&buf, 5)
	gzipW.Write(body)
	gzipW.Close()

	assert.Equal(t, recorder.Body.Len(), buf.Len())
}

func TestCompressGzipPanic(t *testing.T) {
	c := CompressPlugin{}
	configStr := `
    gzip:
       level: 56
`
	var config interface{}
	yaml.Unmarshal([]byte(configStr), &config)

	c.configure(config)
	req, _ := http.NewRequest("GET", "http://mort/local/small.jpg-m", nil)
	req.Header.Add("Accept-Encoding", "gzip")
	body := make([]byte, 1200)
	body[33] = 'a'
	body[324] = 'c'
	res := response.NewBuf(200, body)
	res.Headers.Add("Content-Type", "text/html")

	assert.Panics(t, func() {
		c.postProcess(nil, req, res)

		recorder := httptest.NewRecorder()
		res.Send(recorder)
	})
}

func TestCompressGzipType(t *testing.T) {
	c := CompressPlugin{}
	configStr := `
    gzip:
       level: 5
       types: ["application/json"]
`
	var config interface{}
	yaml.Unmarshal([]byte(configStr), &config)

	c.configure(config)
	req, _ := http.NewRequest("GET", "http://mort/local/small.jpg-m", nil)
	req.Header.Add("Accept-Encoding", "gzip")
	body := make([]byte, 1200)
	body[33] = 'a'
	body[324] = 'c'
	res := response.NewBuf(200, body)
	res.Headers.Add("Content-Type", "application/json")

	c.postProcess(nil, req, res)

	assert.Equal(t, len(res.Headers), 3)
	assert.Equal(t, res.Headers.Get("Content-Encoding"), "gzip")
	assert.Equal(t, res.Headers.Get("Vary"), "Accept-Encoding")

	recorder := httptest.NewRecorder()
	res.Send(recorder)

	var buf bytes.Buffer
	gzipW, _ := gzip.NewWriterLevel(&buf, 5)
	gzipW.Write(body)
	gzipW.Close()

	assert.Equal(t, recorder.Body.Len(), buf.Len())
}

func TestNotCompressOnRange(t *testing.T) {
	c := CompressPlugin{}
	configStr := `
    gzip:
       level: 5
       types: ["application/json"]
`
	var config interface{}
	yaml.Unmarshal([]byte(configStr), &config)

	c.configure(config)
	req, _ := http.NewRequest("GET", "http://mort/local/small.jpg-m", nil)
	req.Header.Add("Accept-Encoding", "gzip")
	req.Header.Add("Range", "0-1")
	body := make([]byte, 1200)
	body[33] = 'a'
	body[324] = 'c'
	res := response.NewBuf(200, body)
	res.Headers.Add("Content-Type", "application/json")

	c.postProcess(nil, req, res)

	assert.Equal(t, len(res.Headers), 1)
}

func TestCompressBrotliType(t *testing.T) {
	c := CompressPlugin{}
	configStr := `
    brotli:
       types: ["application/json"]
`
	var config interface{}
	yaml.Unmarshal([]byte(configStr), &config)

	c.configure(config)
	req, _ := http.NewRequest("GET", "http://mort/local/small.jpg-m", nil)
	req.Header.Add("Accept-Encoding", "gzip, br")
	body := make([]byte, 1200)
	body[33] = 'a'
	body[324] = 'c'
	res := response.NewBuf(200, body)
	res.Headers.Add("Content-Type", "application/json")

	c.postProcess(nil, req, res)

	assert.Equal(t, len(res.Headers), 3)
	assert.Equal(t, res.Headers.Get("Content-Encoding"), "br")
	assert.Equal(t, res.Headers.Get("Vary"), "Accept-Encoding")

	recorder := httptest.NewRecorder()
	res.Send(recorder)

	var buf bytes.Buffer

	br := brEnc.NewWriter(&buf, brEnc.WriterOptions{Quality:4})
	br.Write(body)
	br.Close()

	assert.Equal(t, recorder.Body.Len(), buf.Len())
}

func TestCompressBrImage(t *testing.T) {
	c := CompressPlugin{}
	configStr := `
    gzip:
       level: 5
    brotli:
       types: ["application/json", "text/html"]
`
	var config interface{}
	yaml.Unmarshal([]byte(configStr), &config)

	c.configure(config)
	req, _ := http.NewRequest("GET", "http://mort/local/small.jpg-m", nil)
	req.Header.Add("Accept-Encoding", "gzip, br")
	res := response.NewBuf(200, make([]byte, 13000))
	res.Headers.Add("Content-Type", "image/jpg")

	c.postProcess(nil, req, res)

	assert.Equal(t, len(res.Headers), 1)
}

func TestCompressDoBrImage(t *testing.T) {
	c := CompressPlugin{}
	configStr := `
    gzip:
       level: 5
    brotli:
       types: ["application/json", "text/html"]
`
	var config interface{}
	yaml.Unmarshal([]byte(configStr), &config)

	c.configure(config)
	req, _ := http.NewRequest("GET", "http://mort/local/small.jpg-m", nil)
	req.Header.Add("Accept-Encoding", "gzip, br")
	res := response.NewBuf(200, make([]byte, 13000))
	res.Headers.Add("Content-Type", "text/html")

	c.postProcess(nil, req, res)

	assert.Equal(t, len(res.Headers), 3)
	assert.Equal(t, res.Headers.Get("Content-Encoding"), "br")
}
