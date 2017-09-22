package imgserver


import (

	"imgserver/storage"
	"imgserver/object"
	"imgserver/response"
	"gopkg.in/h2non/bimg.v1"
	"gopkg.in/h2non/filetype.v1"
	"imgserver/engine"
	"net"
	"imgserver/transforms"
)

func Process(obj *object.FileObject) (response.Response) {
	var parent response.Response
	if obj.HasParent() {
		parent = storage.Get(obj.GetParent())
	}

	if len(obj.Transforms) == 0  &&  obj.Params == nil {
		return storage.Get(obj)
	}

	if parent.StatusCode == 404 {
		return parent
	}

	if len(parent.Body) > 261 && filetype.IsImage(parent.Body[:261]) {
		return processImage(parent.Body, obj.Transforms, obj.Params)
	}

	return storage.Get(obj)
}


func processImage(body []byte, transforms transforms.Base, param transforms.Param) (response.Response)
	var result byte[]
	engine := engine.NewImageEngine(body, &resutl}
	for trans := range transforms {
		err := engine.Process(trans)
		if err != nil {
			log.Fatal("Unable to process %s err %s", trans.Name, err)
			return response.WrapWithError
		}
	}

	engine.Param(param)

	result := engine.Result()



}
