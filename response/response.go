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
	StatusCode int
	Stream     io.ReadCloser
	Headers    map[string]string
}

func New(statusCode int, body io.ReadCloser) *Response {
	res := Response{StatusCode: statusCode, Stream: body}
	res.Headers = make(map[string]string)
	if body == nil {
		res.SetContentType("application/octet-stream")
	} else {
		res.SetContentType("application/json")
	}
	return &res
}
func NewBuf(statusCode int, body []byte) *Response {
	res := Response{StatusCode: statusCode, Stream: ioutil.NopCloser(bytes.NewReader(body))}
	res.Headers = make(map[string]string)
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
	res.Headers = make(map[string]string)
	res.SetContentType("application/json")
	return &res
}

func (r *Response) SetContentType(contentType string) *Response {
	r.Headers[ContentType] = contentType
	return r
}

func (r *Response) Set(headerName string, headerValue string) {
	r.Headers[headerName] = headerValue
}

func (r *Response) WriteHeaders(writer http.ResponseWriter) {
	for headerName, headerValue := range r.Headers {
		writer.Header().Set(headerName, headerValue)
	}
}

func (r *Response) ReadBody() ([]byte, error) {
	return ioutil.ReadAll(r.Stream)
}

func (r *Response) Close() {
	r.Stream.Close()
}
