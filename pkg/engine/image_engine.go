package engine

import (
	"encoding/binary"
	"net/http"
	"strconv"
	"time"

	"github.com/spaolacci/murmur3"
	"gopkg.in/h2non/bimg.v1"

	"bytes"
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

	var transHash uint64
	for _, tran := range trans {
		image := bimg.NewImage(buf)
		meta, err := image.Metadata()
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
	res.Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
	res.Set("ETag", createWeakEtag(strconv.FormatUint(hash.Sum64(), 16)))
	meta, err := bimg.Metadata(buf)
	if err == nil {
		res.Set("x-amz-meta-public-width", strconv.Itoa(meta.Size.Width))
		res.Set("x-amz-meta-public-height", strconv.Itoa(meta.Size.Height))

	} else {
		monitoring.Log().Warn("ImageEngine/process unable to get metadata", obj.LogData(zap.Error(err))...)
	}

	return res, nil
}

func createWeakEtag(transHash string) string {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	buf.WriteByte('W')
	buf.WriteByte('/')
	buf.WriteByte('"')
	buf.WriteString(transHash)
	buf.WriteByte('"')
	defer bufPool.Put(buf)
	return buf.String()
}
