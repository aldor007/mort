package engine

import (
	"errors"

	"imgserver/object"
	"imgserver/response"
	Logger "github.com/labstack/gommon/log"
)

type ImageEngine struct {
	parent *response.Response
	Input []byte
}

func NewImageEngine (res *response.Response) *ImageEngine {
	return &ImageEngine{Input: res.Body, parent: res}
}

func (self *ImageEngine) Process(obj *object.FileObject) (*response.Response, error) {
	
	for i, trans := range obj.Transforms {
		Logger.Infof("Permoring i = %d trans = %s", i, trans.Name)
	}

	return response.NewError(400, errors.New("Not ready")), errors.New("Not ready")
}

