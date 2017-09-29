package mort

import (
	"strings"

	"mort/config"
	"mort/engine"
	"mort/object"
	"mort/response"
	"mort/storage"
)

func Process(obj *object.FileObject) *response.Response {
	var parent *response.Response
	if obj.HasParent() {
		parent = storage.Get(obj.GetParent())
	}

	if obj.Transforms.NotEmpty == false {
		return updateHeaders(storage.Get(obj))
	}

	if parent.StatusCode == 404 {
		return updateHeaders(parent)
	}

	if strings.Contains(parent.Headers[response.ContentType], "image/")  {
		return updateHeaders(processImage(parent, obj))
	}

	return updateHeaders(storage.Get(obj))
}

func processImage(parent *response.Response, obj *object.FileObject) *response.Response {
	engine := engine.NewImageEngine(parent)
	result, err := engine.Process(obj)
	if err != nil {
		return response.NewError(400, err)
	}

	return result

}

func updateHeaders(res *response.Response) *response.Response {
	headers := config.GetInstance().Headers
	for _, headerPred := range headers {
		for _, status := range headerPred.StatusCodes {
			if status == res.StatusCode {
				for h, v := range headerPred.Values {
					res.Set(h, v)
				}
				return res
			}
		}
	}
	return res
}
