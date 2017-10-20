package response

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
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
}

func New(statusCode int, body io.ReadCloser) *Response {
	res := Response{StatusCode: statusCode, Stream: body}
	res.Headers = make(http.Header)
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
	body := map[string]string{"message": err.Error()}
	jsonBody, _ := json.Marshal(body)
	res := Response{StatusCode: statusCode, Stream: ioutil.NopCloser(bytes.NewReader(jsonBody))}
	res.ContentLength = int64(len(jsonBody))
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
