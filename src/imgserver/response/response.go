package response


type Response struct{
	StatusCode int
	Body []byte
	Error error
	Headers map[string] string
}

func New(statusCode int, body []byte, err error) *Response{
	res := Response{StatusCode: statusCode, Body: body, Error: err}
	res.Headers = make(map[string] string)
	if body == nil {
		res.SetContentType("application/octet-stream")
	} else {
		res.SetContentType("application/json")
	}
	return &res
}

func (r *Response) SetContentType(contentType string) *Response {
	r.Headers["content-type"] = contentType
	return r
}
