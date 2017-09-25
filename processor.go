package mort

import (
	"gopkg.in/h2non/filetype.v1"
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
		return storage.Get(obj)
	}

	if parent.StatusCode == 404 {
		return parent
	}

	if len(parent.Body) > 261 && filetype.IsImage(parent.Body[:261]) {
		return processImage(parent, obj)
	}

	return storage.Get(obj)
}

func processImage(parent *response.Response, obj *object.FileObject) *response.Response {
	engine := engine.NewImageEngine(parent)
	result, err := engine.Process(obj)
	if err != nil {
		return response.NewError(400, err)
	}

	return result

}
