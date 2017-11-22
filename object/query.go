package object

import (
	"github.com/aldor007/mort/config"
	"github.com/aldor007/mort/transforms"
	"net/url"
	"path"
	"strconv"
)

func decodeQuery(url *url.URL, mortConfig *config.Config, bucketConfig config.Bucket, obj *FileObject) error {
	trans := bucketConfig.Transform

	var err error
	obj.Transforms, err = queryToTransform(url.Query())

	if obj.HasTransform() {
		parentBucket := obj.Bucket
		if trans.ParentBucket != "" {
			parentBucket = trans.ParentBucket
		}

		parent := "/" + path.Join(parentBucket, obj.Key)
		obj.Key = hashKey(obj.Transforms.Hash(), obj.key)
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

	var err error
	opt := query.Get("operation")
	if opt == "" {
		err = trans.Resize(queryToInt(query, "width"), queryToInt(query, "height"), false)
	} else {
		for qsKey, values := range query {
			if qsKey == "operation" {
				for _, o := range values {
					switch o {
					case "resize":
						err = trans.Resize(queryToInt(query, "width"), queryToInt(query, "height"), false)
					case "crop":
						err = trans.Crop(queryToInt(query, "width"), queryToInt(query, "height"), false)
					case "watermark":
						var sigma float64
						sigma, err = strconv.ParseFloat(query.Get("sigma"), 32)
						err = trans.Watermark(query.Get("image"), query.Get("position"), float32(sigma))
					}
				}
			}
		}
	}

	err = trans.Quality(queryToInt(query, "quality"))
	err = trans.Format(query.Get("format"))

	return trans, err
}

func queryToInt(q url.Values, k string) int {
	r, _ := strconv.ParseInt(q.Get(k), 10, 32)
	return int(r)

}
