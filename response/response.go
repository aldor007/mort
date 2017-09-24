package response

import (
	"encoding/json"
	"net/http"
)

const (
	ContentType = "content-type"
)

type Response struct {
	StatusCode int
	Body       []byte
	Headers    map[string]string
}

func New(statusCode int, body []byte) *Response {
	res := Response{StatusCode: statusCode, Body: body}
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
	res := Response{StatusCode: statusCode, Body: jsonBody}
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
