package engine

import (
	"time"
	"strconv"
	"encoding/binary"
	"net/http"

	"gopkg.in/h2non/bimg.v1"
	"github.com/spaolacci/murmur3"

	"mort/transforms"
	"mort/response"
	"mort/object"
	"mort/log"
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

	var transHash uint64
	for _, tran :=  range trans {
		image := bimg.NewImage(buf)
		meta, err := bimg.Metadata(buf)
		if err != nil {
			return response.NewError(500, err), err
		}

		opts, err := tran.BimgOptions(transforms.NewImageInfo(meta, bimg.DetermineImageTypeName(buf)))
		if err != nil {
			return response.NewError(500, err), err
		}

		buf, err = image.Process(opts)
		if err != nil {
			return response.NewError(500, err), err
		}
		transHash = transHash + tran.Hash().Sum64()
	}

	transHashB := make([]byte, 8)
	binary.LittleEndian.PutUint64(transHashB, transHash)

	hash := murmur3.New64()
	hash.Write([]byte(obj.Key))
	hash.Write(transHashB)

	res := response.NewBuf(200, buf)
	res.SetContentType("image/" + bimg.DetermineImageTypeName(buf))
	//res.Set("cache-control", "max-age=6000, public")
	res.Set("Last-Modified", time.Now().Format(http.TimeFormat))
	h := int64(hash.Sum64())

	if h < 0 {
		h = h * -1
	}

	res.Set("ETag", strconv.FormatInt(h, 16))
	meta, err := bimg.Metadata(buf)
	if err == nil {
		res.Set("x-amz-public-width", strconv.Itoa(meta.Size.Width))
		res.Set("x-amz-public-height", strconv.Itoa(meta.Size.Height))

	} else {
		log.Log().Warnw("ImageEngine/process unable to process", "obj.key", obj.Key, "err", err)
	}

	return res, nil
}
