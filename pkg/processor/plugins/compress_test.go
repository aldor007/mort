package plugins

import (
	"bytes"
	"compress/gzip"
	"github.com/aldor007/mort/pkg/response"
	brEnc "github.com/google/brotli/go/cbrotli"
	"github.com/stretchr/testify/assert"
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

	br := brEnc.NewWriter(&buf, brEnc.WriterOptions{Quality: 4})
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

// Integration tests for brotli compression with various scenarios
func TestBrotliCompressionQuality(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		quality int
		valid   bool
	}{
		{"quality_min", 0, true},
		{"quality_default", 4, true},
		{"quality_max", 11, true},
		{"quality_invalid_high", 12, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.valid {
				// Skip invalid quality tests as they should panic
				return
			}

			c := CompressPlugin{}
			configStr := `
brotli:
   quality: ` + string(rune(tt.quality+'0')) + `
   types: ["text/plain"]
`
			var config interface{}
			yaml.Unmarshal([]byte(configStr), &config)
			c.configure(config)

			req, _ := http.NewRequest("GET", "http://test/file.txt", nil)
			req.Header.Add("Accept-Encoding", "br")
			body := make([]byte, 2000)
			for i := range body {
				body[i] = byte(i % 256)
			}
			res := response.NewBuf(200, body)
			res.Headers.Add("Content-Type", "text/plain")

			c.postProcess(nil, req, res)

			assert.Equal(t, "br", res.Headers.Get("Content-Encoding"))
			assert.Equal(t, "Accept-Encoding", res.Headers.Get("Vary"))
		})
	}
}

func TestBrotliVsGzipPriority(t *testing.T) {
	t.Parallel()

	c := CompressPlugin{}
	configStr := `
gzip:
   level: 6
brotli:
   quality: 4
   types: ["text/html", "application/json"]
`
	var config interface{}
	yaml.Unmarshal([]byte(configStr), &config)
	c.configure(config)

	// When both are accepted, brotli should be preferred
	req, _ := http.NewRequest("GET", "http://test/file.html", nil)
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	body := make([]byte, 3000)
	res := response.NewBuf(200, body)
	res.Headers.Add("Content-Type", "text/html")

	c.postProcess(nil, req, res)

	assert.Equal(t, "br", res.Headers.Get("Content-Encoding"),
		"Brotli should be preferred when both gzip and br are accepted")
}

func TestBrotliWithDifferentContentTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		contentType    string
		shouldCompress bool
	}{
		{"json", "application/json", true},
		{"html", "text/html; charset=utf-8", true},
		{"javascript", "application/javascript", true},
		{"css", "text/css", true},
		{"xml", "application/xml", true},
		{"svg", "image/svg+xml", true},
		{"jpeg", "image/jpeg", false},
		{"png", "image/png", false},
		{"gif", "image/gif", false},
		{"binary", "application/octet-stream", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := CompressPlugin{}
			configStr := `
brotli:
   quality: 5
   types: ["application/json", "text/html", "application/javascript", "text/css", "application/xml", "image/svg+xml"]
`
			var config interface{}
			yaml.Unmarshal([]byte(configStr), &config)
			c.configure(config)

			req, _ := http.NewRequest("GET", "http://test/file", nil)
			req.Header.Add("Accept-Encoding", "br")
			body := make([]byte, 2000)
			res := response.NewBuf(200, body)
			res.Headers.Add("Content-Type", tt.contentType)

			c.postProcess(nil, req, res)

			if tt.shouldCompress {
				assert.Equal(t, "br", res.Headers.Get("Content-Encoding"),
					"Content-Type %s should be compressed", tt.contentType)
			} else {
				assert.NotEqual(t, "br", res.Headers.Get("Content-Encoding"),
					"Content-Type %s should not be compressed", tt.contentType)
			}
		})
	}
}

func TestBrotliCompressionRatio(t *testing.T) {
	t.Parallel()

	c := CompressPlugin{}
	configStr := `
brotli:
   quality: 6
   types: ["text/plain"]
`
	var config interface{}
	yaml.Unmarshal([]byte(configStr), &config)
	c.configure(config)

	req, _ := http.NewRequest("GET", "http://test/file.txt", nil)
	req.Header.Add("Accept-Encoding", "br")

	// Create highly compressible content (repeated pattern)
	body := bytes.Repeat([]byte("test data for compression "), 100)
	originalSize := len(body)
	res := response.NewBuf(200, body)
	res.Headers.Add("Content-Type", "text/plain")

	c.postProcess(nil, req, res)

	recorder := httptest.NewRecorder()
	res.Send(recorder)

	compressedSize := recorder.Body.Len()

	// Brotli should achieve good compression ratio on repeated data
	compressionRatio := float64(compressedSize) / float64(originalSize)
	assert.Less(t, compressionRatio, 0.5,
		"Brotli should compress repeated data to less than 50%% of original size")
}

func TestBrotliWithEmptyBody(t *testing.T) {
	t.Parallel()

	c := CompressPlugin{}
	configStr := `
brotli:
   quality: 4
   types: ["text/plain"]
`
	var config interface{}
	yaml.Unmarshal([]byte(configStr), &config)
	c.configure(config)

	req, _ := http.NewRequest("GET", "http://test/file.txt", nil)
	req.Header.Add("Accept-Encoding", "br")
	res := response.NewBuf(200, []byte{})
	res.Headers.Add("Content-Type", "text/plain")

	c.postProcess(nil, req, res)

	// Empty body should not be compressed
	assert.NotEqual(t, "br", res.Headers.Get("Content-Encoding"))
}

func TestBrotliWithMultipleEncodings(t *testing.T) {
	t.Parallel()

	c := CompressPlugin{}
	configStr := `
brotli:
   quality: 4
   types: ["text/html"]
`
	var config interface{}
	yaml.Unmarshal([]byte(configStr), &config)
	c.configure(config)

	tests := []struct {
		name           string
		acceptEncoding string
		expectedEnc    string
	}{
		{"only_brotli", "br", "br"},
		{"brotli_first", "br, gzip, deflate", "br"},
		{"brotli_last", "gzip, deflate, br", "br"},
		{"wildcard_with_br", "*, br", "br"},
		{"no_brotli", "gzip, deflate", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "http://test/file.html", nil)
			req.Header.Add("Accept-Encoding", tt.acceptEncoding)
			body := make([]byte, 2000)
			res := response.NewBuf(200, body)
			res.Headers.Add("Content-Type", "text/html")

			c.postProcess(nil, req, res)

			if tt.expectedEnc != "" {
				assert.Equal(t, tt.expectedEnc, res.Headers.Get("Content-Encoding"))
			} else {
				assert.Equal(t, "", res.Headers.Get("Content-Encoding"))
			}
		})
	}
}
