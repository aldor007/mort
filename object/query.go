package object

import (
	"github.com/aldor007/mort/config"
	"github.com/aldor007/mort/transforms"
	"net/url"
	"path"
	"strconv"
	"strings"
)

func decodeQuery(url *url.URL, mortConfig *config.Config, bucketConfig config.Bucket, obj *FileObject) error {
	trans := bucketConfig.Transform
	parent := obj.Key

	var err error
	obj.Transforms, err = queryToTransform(url.Query())

	parent = "/" + path.Join(trans.ParentBucket, parent)

	if obj.HasTransform() {
		obj.Key = "/" + strings.Join([]string{strconv.FormatUint(uint64(obj.Transforms.Hash().Sum64()), 16), parent}, "-")
		parentObj, err := NewFileObjectFromPath(parent, mortConfig)
		parentObj.Storage = bucketConfig.Storages.Get(trans.ParentStorage)
		obj.Parent = parentObj
		obj.CheckParent = trans.CheckParent
		return err
	}

	return err
}

func queryToTransform(query url.Values) (transforms.Transforms, error) {
	var trans transforms.Transforms
	if len(query) == 0 {
		return trans, nil
	}

	opt := query.Get("operation")
	if opt == "" {
		opt = "resize"
	}

	var err error
	switch opt {
	case "resize":
		err = trans.Resize(queryToInt(query, "width"), queryToInt(query, "height"), false)
	case "crop":
		err = trans.Crop(queryToInt(query, "width"), queryToInt(query, "height"), false)
	}
	//case "watermark":
	//opacity, err := strnconv. query.Get("opacity")
	//err = trans.Watermark(query.Get("image"), query.Get("position"),
	err = trans.Quality(queryToInt(query, "quality"))

	return trans, err
}

func queryToInt(q url.Values, k string) int {
	r, _ := strconv.ParseInt(q.Get(k), 10, 32)
	return int(r)

}
