package imgserver


import (

	"imgserver/storage"
	"imgserver/object"
	"imgserver/response"
	"gopkg.in/h2non/bimg.v1"
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

func processResponse(obj *object.FileObject, res *response.Response) (*response.Response) {
	processImage(obj, res)
}

func processImage(obj   *object.FileObject, res *response.Response) {
	bimg.
}
