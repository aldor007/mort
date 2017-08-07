package imgserver


import (
	"imgserver/storage"
	"imgserver/object"
	"imgserver/response"
)

func Process(obj *object.FileObject) (*response.Response) {
	// check if parent exists
	if obj.HasParent() {
		// TODO head method
		parent := storage.Get(obj.GetParent())
		if parent.StatusCode == 404 {
			return parent
		}
	}

	res := storage.Get(obj)

	return res

}