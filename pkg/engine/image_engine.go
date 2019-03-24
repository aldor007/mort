package engine

import (
	"net/http"
	"strconv"
	"time"

	"gopkg.in/h2non/bimg.v1"

	"bytes"
	"crypto/md5"
	"encoding/hex"
	"github.com/aldor007/mort/pkg/monitoring"
	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/response"
	"github.com/aldor007/mort/pkg/transforms"
	"go.uber.org/zap"
	"sync"
)

// bufPool for string concatenations
var bufPool = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

// ImageEngine is main struct that is responding for image processing
type ImageEngine struct {
	parent *response.Response // source file
}

// NewImageEngine create instance of ImageEngine with source file that should be processed
func NewImageEngine(res *response.Response) *ImageEngine {
	return &ImageEngine{parent: res}
}

// Process main ImageEngine function that create new image (stored in response object)
func (c *ImageEngine) Process(obj *object.FileObject, trans []transforms.Transforms) (*response.Response, error) {
	t := monitoring.Report().Timer("generation_time")
	defer t.Done()

	buf, err := c.parent.ReadBody()
	if err != nil {
		return response.NewError(500, err), err
	}

	for _, tran := range trans {
		image := bimg.NewImage(buf)
		meta, err := image.Metadata()
		if err != nil {
			return response.NewError(500, err), err
		}

		optsArr, err := tran.BimgOptions(transforms.NewImageInfo(meta, bimg.DetermineImageTypeName(buf)))
		if err != nil {
			return response.NewError(500, err), err
		}
		optsLen := len(optsArr)
		for i, opts := range optsArr {
			buf, err = image.Process(opts)
			if err != nil {
				return response.NewError(500, err), err
			}

			if i <= optsLen-1 {
				image = bimg.NewImage(buf)
			}
		}
	}

	bodyHash := md5.New()
	bodyHash.Write(buf)

	res := response.NewBuf(200, buf)
	res.SetContentType("image/" + bimg.DetermineImageTypeName(buf))
	//res.Set("cache-control", "max-age=6000, public")
	res.Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
	res.Set("ETag", hex.EncodeToString(bodyHash.Sum(nil)))
	meta, err := bimg.Metadata(buf)
	if err == nil {
		res.Set("x-amz-meta-public-width", strconv.Itoa(meta.Size.Width))
		res.Set("x-amz-meta-public-height", strconv.Itoa(meta.Size.Height))

	} else {
		monitoring.Log().Warn("ImageEngine/process unable to get metadata", obj.LogData(zap.Error(err))...)
	}

	return res, nil
}
