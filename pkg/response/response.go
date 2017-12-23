package response

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/aldor007/mort/pkg/log"
	"github.com/aldor007/mort/pkg/object"
	"github.com/djherbis/stream"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	// HeaderContentType name of Content-Type header
	HeaderContentType = "content-type"
)

func isRangeOrCondition(req *http.Request) bool {
	if req.Header.Get("Range") != "" || req.Header.Get("if-range") != "" {
		return true
	}

	if req.Header.Get("If-match") != "" || req.Header.Get("If-none-match") != "" {
		return true
	}

	if req.Header.Get("If-Unmodified-Since") != "" || req.Header.Get("If-Modified-Since") != "" {
		return true
	}

	return false
}

// Response is helper struct for wrapping different storage response
type Response struct {
	StatusCode    int         // status code of response
	Headers       http.Header // headers for response
	ContentLength int64       // if buffered response contains length of buffer, for streams it equal to -1
	debug         bool        // debug flag
	errorValue    error       // error value

	reader     io.ReadCloser // reader for response body
	body       []byte        // response body for buffered value
	bodyReader io.ReadCloser // original response buffer
	bodySeeker io.ReadSeeker

	resStream *stream.Stream // response stream dispatcher
	hasParent bool           // flag indicated that response is a copy
}

// New create response object with io.ReadCloser
func New(statusCode int, body io.ReadCloser) *Response {
	res := Response{StatusCode: statusCode, reader: body}
	res.ContentLength = 0
	if body != nil {
		seeker, ok := body.(io.ReadSeeker)
		if ok {
			res.bodySeeker = seeker
		}
		res.ContentLength = -1
	}
	res.Headers = make(http.Header)
	return &res
}

// NewNoContent create response object without content
func NewNoContent(statusCode int) *Response {
	res := New(statusCode, nil)
	res.ContentLength = 0
	return res
}

// NewString create response object from string
func NewString(statusCode int, body string) *Response {
	res := Response{StatusCode: statusCode}
	res.bodySeeker = strings.NewReader(body)
	res.reader = ioutil.NopCloser(res.bodySeeker)
	res.ContentLength = int64(len(body))
	res.Headers = make(http.Header)
	res.Headers.Set(HeaderContentType, "text/plain")
	return &res
}

// NewBuf create response object from []byte
func NewBuf(statusCode int, body []byte) *Response {
	res := Response{StatusCode: statusCode}
	res.bodySeeker = bytes.NewReader(body)
	res.reader = ioutil.NopCloser(res.bodySeeker)
	res.ContentLength = int64(len(body))
	res.body = body
	res.Headers = make(http.Header)
	return &res
}

// NewError create response object from error
func NewError(statusCode int, err error) *Response {
	res := Response{StatusCode: statusCode, errorValue: err}
	res.Headers = make(http.Header)
	res.Headers.Set(HeaderContentType, "application/json")
	return &res
}

// SetContentType update content type header of response
func (r *Response) SetContentType(contentType string) *Response {
	r.Headers.Set(HeaderContentType, contentType)
	return r
}

// Set update response headers
func (r *Response) Set(headerName string, headerValue string) {
	r.Headers.Set(headerName, headerValue)
}

// ReadBody reads all content of response and returns []byte
// Content of the response is changed
// Such response shouldn't be Send to client
func (r *Response) ReadBody() ([]byte, error) {
	if r.body != nil {
		return r.body, nil
	}

	if r.reader == nil {
		return nil, errors.New("empty body")
	}

	body, err := ioutil.ReadAll(r.reader)
	r.body = body
	return r.body, err
}

// CopyBody read all content of response and returns it in []byte
// but doesn't change response object body
func (r *Response) CopyBody() ([]byte, error) {
	var buf []byte
	if r.body != nil {
		buf = r.body
	} else {
		var err error
		buf, err = ioutil.ReadAll(r.reader)

		if err != nil {
			return nil, err
		}

		r.reader.Close()
		r.body = buf
		r.reader = ioutil.NopCloser(bytes.NewReader(buf))
	}

	return r.body, nil
}

// Close response reader
func (r *Response) Close() {
	if r.reader != nil {
		r.reader.Close()
		r.reader = nil
	}

	if r.bodyReader != nil {
		r.bodyReader.Close()
		r.bodyReader = nil
	}

	if r.resStream != nil && r.hasParent == false {
		go func() {
			r.resStream.Close()
			r.resStream.Remove()
		}()
	}
}

// SetDebug set flag indicating that response can including debug information
func (r *Response) SetDebug(debug bool, obj *object.FileObject) *Response {
	if debug == true {
		r.debug = true
		if obj != nil {
			r.Headers.Set("Cache-Control", "no-cache")
			r.Headers.Set("x-mort-key", obj.Key)
			r.Headers.Set("x-mort-storage", obj.Storage.Kind)

			if obj.HasTransform() {
				r.Headers.Set("x-mort-transform", "true")
			}

			if obj.HasParent() {
				r.Headers.Set("x-mort-parent-key", obj.Parent.Key)
				r.Headers.Set("x-mort-parent-bucket", obj.Parent.Bucket)
				r.Headers.Set("x-mort-parent-storage", obj.Parent.Storage.Kind)
			}
		}
		r.writeDebug()
		return r
	}

	r.debug = false
	return r
}

// HasError check if response contains error
func (r *Response) HasError() bool {
	return r.errorValue != nil
}

// Error returns error instance
func (r *Response) Error() error {
	return r.errorValue
}

// Send write response to client using streaming
func (r *Response) Send(w http.ResponseWriter) error {
	for headerName, headerValue := range r.Headers {
		w.Header().Set(headerName, headerValue[0])
	}

	defer r.Close()
	var resStream io.Reader
	if r.ContentLength != 0 {
		resStream = r.Stream()
		if resStream == nil {
			r.StatusCode = 500
		}
	}

	w.WriteHeader(r.StatusCode)

	if resStream != nil {
		io.Copy(w, resStream)
	}

	return nil
}

// SendContent use http.ServeContent to return response to client
// It can handle range and condition requests
func (r *Response) SendContent(req *http.Request, w http.ResponseWriter) error {
	if r.StatusCode != 200 || r.bodySeeker == nil || isRangeOrCondition(req) == false {
		log.Log().Info("Response SendContent streaming response", zap.Int("sc", r.StatusCode), zap.Bool("bodySeeker", r.bodySeeker != nil), zap.Bool("rangeOrConditionReq", isRangeOrCondition(req)))
		return r.Send(w)
	}

	defer r.Close()
	for headerName, headerValue := range r.Headers {
		w.Header().Set(headerName, headerValue[0])
	}

	lastMod, err := time.Parse(http.TimeFormat, r.Headers.Get("Last-Modified"))
	if err != nil {
		lastMod = time.Now()
	}

	http.ServeContent(w, req, "", lastMod, r.bodySeeker)

	return nil
}

// CopyHeadersFrom copy all headers from src response but body is omitted
func (r *Response) CopyHeadersFrom(src *Response) {
	r.Headers = make(http.Header, len(src.Headers))
	for k, v := range src.Headers {
		r.Headers[k] = v
	}

	r.StatusCode = src.StatusCode
	r.ContentLength = src.ContentLength
	r.debug = src.debug
	r.errorValue = src.errorValue
}

// Copy create complete response copy with headers and body
func (r *Response) Copy() (*Response, error) {
	if r == nil {
		return nil, nil
	}

	c := Response{StatusCode: r.StatusCode, ContentLength: r.ContentLength, debug: r.debug, errorValue: r.errorValue}
	c.Headers = make(http.Header)
	for k, v := range r.Headers {
		c.Headers[k] = v
	}

	if r.body != nil {
		c.ContentLength = int64(len(r.body))
		c.body = r.body
		c.bodySeeker = bytes.NewReader(c.body)
	} else if r.reader != nil {
		buf, err := r.CopyBody()
		if err != nil {
			return nil, err
		}

		c.bodySeeker = bytes.NewReader(buf)
		c.reader = ioutil.NopCloser(c.bodySeeker)
		c.ContentLength = int64(len(buf))
		c.body = buf

	}

	return &c, nil

}

// CopyWithStream should be used with not buffered response that contain stream
// it try duplicate response stream for multiple readers
func (r *Response) CopyWithStream() (*Response, error) {
	if r.body != nil {
		return r.Copy()
	}

	c := Response{StatusCode: r.StatusCode, ContentLength: r.ContentLength, debug: r.debug, errorValue: r.errorValue}
	c.Headers = make(http.Header)
	for k, v := range r.Headers {
		c.Headers[k] = v
	}

	if r.resStream != nil {
		c.resStream = r.resStream
		c.hasParent = true
		return &c, nil
	}

	r.bodyReader = r.reader

	r.resStream = stream.NewMemStream()
	c.resStream = r.resStream
	c.hasParent = true
	r.reader = ioutil.NopCloser(io.TeeReader(r.bodyReader, r.resStream))

	return &c, nil

}

// Stream return io.Reader interferace from correct response content
func (r *Response) Stream() io.ReadCloser {
	if r.hasParent == true && r.resStream != nil {
		r, _ := r.resStream.NextReader()
		return ioutil.NopCloser(r)
	}

	if r.body != nil {
		return ioutil.NopCloser(bytes.NewReader(r.body))
	}

	if r.reader != nil {
		return r.reader
	}

	return nil
}

// IsBuffered check if response has access to original buffor
func (r *Response) IsBuffered() bool {
	return r.body != nil
}

func (r *Response) writeDebug() {
	if !r.debug {
		return
	}

	if r.errorValue != nil {

		body := map[string]string{"message": r.errorValue.Error()}
		jsonBody, err := json.Marshal(body)
		if err != nil {
			panic(err)
		}
		r.reader = ioutil.NopCloser(bytes.NewReader(jsonBody))
		r.body = jsonBody
		r.ContentLength = int64(len(jsonBody))
		r.SetContentType("application/json")
	}
}
