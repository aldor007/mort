package engine

import (

	"mort/object"
	"mort/response"
	//Logger "github.com/labstack/gommon/log"
	"gopkg.in/h2non/bimg.v1"
)

type ImageEngine struct {
	Input []byte
	parent *response.Response
}

func NewImageEngine (res *response.Response) *ImageEngine {
	return &ImageEngine{Input: res.Body, parent: res}
}

func (self *ImageEngine) Process(obj *object.FileObject) (*response.Response, error) {

	image := bimg.NewImage(self.Input)
	buf, err := image.Process(obj.Transforms.BimgOptions())
	if err != nil  {
		return response.NewError(500, err), err
	}

	res := response.New(200, buf)
	res.SetContentType("image/" + bimg.DetermineImageTypeName(buf))
	res.Set("cache-control", "max-age=6000, public")

	return res, nil
}

