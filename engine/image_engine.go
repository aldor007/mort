package engine

import (
	"mort/object"
	"mort/response"
	//Logger "github.com/labstack/gommon/log"
	"gopkg.in/h2non/bimg.v1"
)

type ImageEngine struct {
	parent *response.Response
}

func NewImageEngine(res *response.Response) *ImageEngine {

	return &ImageEngine{parent: res}
}

func (self *ImageEngine) Process(obj *object.FileObject) (*response.Response, error) {
	body, err := self.parent.ReadBody()
	if err != nil {
		return response.NewError(500, err), err
	}

	image := bimg.NewImage(body)
	buf, errBody := image.Process(obj.Transforms.BimgOptions())
	if errBody != nil {
		return response.NewError(500, errBody), errBody
	}

	res := response.NewBuf(200, buf)
	res.SetContentType("image/" + bimg.DetermineImageTypeName(buf))
	res.Set("cache-control", "max-age=6000, public")

	return res, nil
}
