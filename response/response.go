package response

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"github.com/labstack/echo"
)

const (
	ContentType = "content-type"
)

type Response struct {
	StatusCode    int
	Stream        io.ReadCloser
	Headers       http.Header
	ContentLength int64
	ContentType   string
	debug 		  bool
	errorValue    error
	lock          sync.RWMutex
}

func New(statusCode int, body io.ReadCloser) *Response {
	res := Response{StatusCode: statusCode, Stream: body}
	res.Headers = make(http.Header)
	res.ContentLength = -1
	if body == nil {
		res.SetContentType("application/octet-stream")
	} else {
		res.SetContentType("application/json")
	}
	return &res
}
func NewBuf(statusCode int, body []byte) *Response {
	res := Response{StatusCode: statusCode, Stream: ioutil.NopCloser(bytes.NewReader(body))}
	res.ContentLength = int64(len(body))
	res.Headers = make(http.Header)
	if body == nil {
		res.SetContentType("application/octet-stream")
	} else {
		res.SetContentType("application/json")
	}
	return &res
}

func NewError(statusCode int, err error) *Response {
	res := Response{StatusCode: statusCode, errorValue: err}
	res.Headers = make(http.Header)
	res.SetContentType("application/json")
	return &res
}

func (r *Response) SetContentType(contentType string) *Response {
	r.Headers.Set(ContentType, contentType)
	r.ContentType = contentType
	return r
}

func (r *Response) Set(headerName string, headerValue string) {
	r.Headers.Set(headerName, headerValue)
}

func (r *Response) WriteHeaders(writer http.ResponseWriter) {
	r.writeDebug()

	for headerName, headerValue := range r.Headers {
		writer.Header().Set(headerName, headerValue[0])
	}

	writer.Header().Set(ContentType, r.ContentType)
}

func (r *Response) ReadBody() ([]byte, error) {
	return ioutil.ReadAll(r.Stream)
}

func (r *Response) CopyBody() ([]byte, error) {
	buf, err := ioutil.ReadAll(r.Stream)
	if err != nil {
		return nil, err
	}

	r.Stream.Close()
	r.Stream = ioutil.NopCloser(bytes.NewReader(buf))
	return buf, nil
}

func (r *Response) Close() {
	if r.Stream != nil {
		r.Stream.Close()
	}
}

func (r *Response) SetDebug(debug string)  {
	if debug == "1" {
		r.debug = true
		r.Set("Cache-Control",  "no-cache")
		return
	}

	r.debug = false
}

func (r *Response) HasError()  bool {
	return r.errorValue != nil
}

func (r *Response) Error()  error {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.errorValue
}

func (r *Response) Write(ctx echo.Context) error {
	if r.ContentLength != 0 {

		defer r.Close()
		return ctx.Stream(r.StatusCode, r.ContentType, r.Stream)
	}

	return ctx.NoContent(r.StatusCode)
}

func (r * Response) Copy() (*Response, error) {
	if r == nil {
		return nil, nil
	}

	c := Response{StatusCode:r.StatusCode, ContentType:r.ContentType,ContentLength: r.ContentLength, debug: r.debug, errorValue:r.errorValue}
	c.Headers = make(http.Header)
	for k, v := range r.Headers {
		c.Headers[k] = v
	}

	buf, err := r.CopyBody()
	if err != nil {
		return nil, err
	}

	c.Stream =  ioutil.NopCloser(bytes.NewReader(buf))
	c.ContentLength = int64(len(buf))

	return &c, nil

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
		r.Stream = ioutil.NopCloser(bytes.NewReader(jsonBody))
		r.ContentLength = int64(len(jsonBody))
		r.SetContentType("application/json")
	}
}