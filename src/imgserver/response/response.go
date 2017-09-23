package response

import  "encoding/json"

const (
	ContentType = "content-type"
)

type Response struct{
	StatusCode int
	Body []byte
	Headers map[string] string
}

func New(statusCode int, body []byte) *Response{
	res := Response{StatusCode: statusCode, Body: body}
	res.Headers = make(map[string] string)
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
	res.Headers = make(map[string] string)
	res.SetContentType("application/json")
	return &res
}

func (r *Response) SetContentType(contentType string) *Response {
	r.Headers[ContentType] = contentType
	return r
}
