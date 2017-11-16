package response

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"errors"
	"github.com/djherbis/stream"
)

const (
	HeaderContentType = "content-type"
)

// Response is helper struct for wrapping diffrent storage response
type Response struct {
	StatusCode    int
	Headers       http.Header
	ContentLength int64
	debug 		  bool
	errorValue    error
	errorWritten  bool

	reader        io.ReadCloser
	body          []byte
	bodyReader     io.ReadCloser
	resStream     *stream.Stream
	parent        *Response
}

// New create response object with io.ReadCloser
func New(statusCode int, body io.ReadCloser) *Response {
	res := Response{StatusCode: statusCode, reader: body}
	res.Headers = make(http.Header)
	res.ContentLength = -1
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
	r := New(statusCode, ioutil.NopCloser(strings.NewReader(body)))
	r.Headers.Set(HeaderContentType, "text/plain")
	return r
}

// NewString create response object from []byte
func NewBuf(statusCode int, body []byte) *Response {
	res := Response{StatusCode: statusCode, reader: ioutil.NopCloser(bytes.NewReader(body))}
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
func (r *Response) Set(headerName string, headerValue string)  {
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
	}

	if r.bodyReader != nil {
		r.bodyReader.Close()
	}

	if r.resStream != nil && r.parent == nil {
		go func() {
			r.resStream.Close()
			//r.resStream.Remove()
		}()
	}
}

// SetDebug set flag indicating that response can including debug information
func (r *Response) SetDebug(debug bool) (*Response) {
	if debug == true {
		r.debug = true
		r.Headers.Set("Cache-Control", "no-cache")
		r.writeDebug()
		return r
	}

	r.debug = false
	return r
}

// HasError check if response contains error
func (r *Response) HasError()  bool {
	return r.errorValue != nil
}

// Error returns error instance
func (r *Response) Error()  error {
	return r.errorValue
}

// Send write response to client
func (r *Response) Send(w http.ResponseWriter) error {
	for headerName, headerValue := range r.Headers {
		w.Header().Set(headerName, headerValue[0])
	}

	w.WriteHeader(r.StatusCode)

	if r.ContentLength != 0 {

		defer r.Close()
		_, err := io.Copy(w, r.Stream())
		return err
	}

	return nil
}

// CopyHeadersFrom copy all headers from src response but body is omitted
func (r * Response) CopyHeadersFrom(src *Response)  {
	r.Headers = make(http.Header, len(src.Headers))
	for k, v := range src.Headers {
		r.Headers[k] = v
	}

	r.StatusCode = src.StatusCode
	r.ContentLength = src.ContentLength
	r.debug = src.debug
	r.errorValue = src.errorValue
}

// Copy create copmlete response copy with headers and body
func (r * Response) Copy() (*Response, error) {
	if r == nil {
		return nil, nil
	}

	c := Response{StatusCode:r.StatusCode, ContentLength: r.ContentLength, debug: r.debug, errorValue:r.errorValue}
	c.Headers = make(http.Header)
	for k, v := range r.Headers {
		c.Headers[k] = v
	}

	if (r.body != nil) {
		c.ContentLength = int64(len(r.body))
		c.body = r.body
	} else if r.reader != nil {
		buf, err := r.CopyBody()
		if err != nil {
			return nil, err
		}

		c.reader =  ioutil.NopCloser(bytes.NewReader(buf))
		c.ContentLength = int64(len(buf))
		c.body = buf

	}

	return &c, nil

}

func (r *Response) CopyWithStream() (*Response, error)  {
	if r.body != nil {
		return r.Copy()
	}

	c := Response{StatusCode:r.StatusCode, ContentLength: r.ContentLength, debug: r.debug, errorValue:r.errorValue}
	c.Headers = make(http.Header)
	for k, v := range r.Headers {
		c.Headers[k] = v
	}

	//c.reader =  ioutil.NopCloser(bytes.NewReader(buf))
	r.bodyReader = r.reader

	r.resStream = stream.NewMemStream()
	c.resStream = r.resStream
	c.parent = r
	r.reader = ioutil.NopCloser(io.TeeReader(r.bodyReader, r.resStream))


	return &c, nil

}

func (r *Response) writeDebug() {
	if !r.debug {
		return
	}

	if r.errorValue != nil  {

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

func (r *Response) Stream() io.Reader {
	if r.parent != nil && r.resStream != nil {
		r, _  := r.resStream.NextReader()
		return r
	}

	if r.reader != nil {
		return r.reader
	}

	if r.body != nil {
		return bytes.NewReader(r.body)
	}

	return nil
}

func (r *Response) IsBuffered() bool {
	return r.body != nil
}

