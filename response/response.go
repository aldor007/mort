package response

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"errors"
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
	errorWritten  bool
}

func New(statusCode int, body io.ReadCloser) *Response {
	res := Response{StatusCode: statusCode, Stream: body}
	res.Headers = make(http.Header)
	if body == nil {
		res.SetContentType("application/octet-stream")
	} else {
		res.SetContentType("application/json")
		res.ContentLength = -1
	}
	return &res
}

func NewNoContent(statusCode int) *Response {
	res := New(statusCode, nil)
	res.ContentLength = 0
	return res
}

func NewString(statusCode int, body string) *Response {
	r := New(statusCode, ioutil.NopCloser(strings.NewReader(body)))
	r.SetContentType("text/plain")
	return r
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

func (r *Response) writeHeaders(writer http.ResponseWriter) {
}

func (r *Response) ReadBody() ([]byte, error) {
	if r.Stream == nil {
		return nil, errors.New("empty body")
	}

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

func (r *Response) SetDebug(debug bool) (*Response) {
	if debug == true {
		r.debug = true
		r.Set("Cache-Control",  "no-cache")
		r.writeDebug()
		return r
	}

	r.debug = false
	return r
}

func (r *Response) HasError()  bool {
	return r.errorValue != nil
}

func (r *Response) Error()  error {
	return r.errorValue
}

func (r *Response) Send(w http.ResponseWriter) error {
	for headerName, headerValue := range r.Headers {
		w.Header().Set(headerName, headerValue[0])
	}

	w.Header().Set(ContentType, r.ContentType)
	w.WriteHeader(r.StatusCode)

	if r.ContentLength != 0 {

		defer r.Close()
		_, err := io.Copy(w, r.Stream)
		return err
	}

	return nil
}


func (r * Response) CopyHeadersFrom(src *Response)  {
	r.Headers = make(http.Header, len(src.Headers))
	for k, v := range src.Headers {
		r.Headers[k] = v
	}

	r.StatusCode = src.StatusCode
	r.ContentType = src.ContentType
	r.ContentLength = src.ContentLength
	r.debug = src.debug
	r.errorValue = src.errorValue
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

	if r.Stream != nil {
		buf, err := r.CopyBody()
		if err != nil {
			return nil, err
		}

		c.Stream =  ioutil.NopCloser(bytes.NewReader(buf))
		c.ContentLength = int64(len(buf))

	}

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
		r.Stream = ioutil.NopCloser(bytes.NewReader(jsonBody))
		r.ContentLength = int64(len(jsonBody))
		r.SetContentType("application/json")
	}
}