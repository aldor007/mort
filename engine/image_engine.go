package engine

import (
	"time"
	"strconv"

	"gopkg.in/h2non/bimg.v1"
	"github.com/spaolacci/murmur3"

	"mort/transforms"
	"mort/response"
	"mort/object"
	//Logger "github.com/labstack/gommon/log"
)

type ImageEngine struct {
	parent *response.Response
}

func NewImageEngine(res *response.Response) *ImageEngine {

	return &ImageEngine{parent: res}
}

func (self *ImageEngine) Process(obj *object.FileObject, trans []transforms.Transforms) (*response.Response, error) {
	buf, err := self.parent.ReadBody()
	if err != nil {
		return response.NewError(500, err), err
	}

	for _, tran :=  range trans {
		image := bimg.NewImage(buf)
		buf, err = image.Process(tran.BimgOptions())
		if err != nil {
			return response.NewError(500, err), err
		}
	}

	hash := murmur3.New32()
	hash.Write([]byte(obj.Key))
	//hash.Write([]byte(len(trans)))

	res := response.NewBuf(200, buf)
	res.SetContentType("image/" + bimg.DetermineImageTypeName(buf))
	res.Set("cache-control", "max-age=6000, public")
	res.Set("last-modified", time.Now().Format(time.RFC1123))
	res.Set("etag", strconv.FormatInt(int64(hash.Sum32()), 16))
	meta, err := bimg.Metadata(buf)
	if err == nil {
		res.Set("x-amz-public-width", strconv.Itoa(meta.Size.Width))
		res.Set("x-amz-public-height", strconv.Itoa(meta.Size.Height))
	}

	return res, nil
}
