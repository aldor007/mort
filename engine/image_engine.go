package engine

import (
	"mort/transforms"
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

func (self *ImageEngine) Process(trans []transforms.Transforms) (*response.Response, error) {
	body, err := self.parent.ReadBody()
	if err != nil {
		return response.NewError(500, err), err
	}
	var buf []byte
	image := bimg.NewImage(body)
	for _, tran :=  range trans {
		buf, err := image.Process(tran.BimgOptions())
		if err != nil {
			return response.NewError(500, err), err
		}
		image = bimg.NewImage(buf)
	}

	res := response.NewBuf(200, buf)
	res.SetContentType("image/" + bimg.DetermineImageTypeName(buf))
	res.Set("cache-control", "max-age=6000, public")

	return res, nil
}
