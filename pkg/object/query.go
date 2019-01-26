package object

import (
	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/transforms"
	"net/url"
	"path"
	"strconv"
)

func init() {
	RegisterParser("query", decodeQuery)
}

func decodeQuery(url *url.URL, bucketConfig config.Bucket, obj *FileObject) (string, error) {
	trans := bucketConfig.Transform

	var err error
	obj.Transforms, err = queryToTransform(url.Query())

	if obj.HasTransform() {
		parent := url.Path
		if trans.ParentBucket != "" {
			parent = "/" + path.Join(trans.ParentBucket, obj.Key)
		}

		return parent, err
	}

	return "", err
}

// nolint: gocyclo
func queryToTransform(query url.Values) (transforms.Transforms, error) {
	if len(query) == 0 {
		return transforms.Transforms{}, nil
	}

	trans, err := parseOperation(query)

	if err != nil {
		return trans, err
	}

	var q int
	if _, ok := query["quality"]; ok {
		q, _ = queryToInt(query, "quality")
		trans.Quality(q)
	}

	if format, ok := query["format"]; ok {
		err = trans.Format(format[0])
		if err != nil {
			return trans, err
		}
	}

	if _, ok := query["grayscale"]; ok {
		trans.Grayscale()
	}

	return trans, err
}

func queryToInt(q url.Values, k string) (int, error) {
	r, err := strconv.ParseInt(q.Get(k), 10, 32)
	return int(r), err

}

func parseOperation(query url.Values) (transforms.Transforms, error) {
	trans := transforms.New()
	var err error
	opt := query.Get("operation")
	if opt == "" {
		w, err1 := queryToInt(query, "width")
		h, err2 := queryToInt(query, "height")
		if (err1 == nil || err2 == nil) && (w != 0 || h != 0) {
			err = trans.Resize(w, h, false)
			if err != nil {
				return trans, err
			}
		}
		return trans, err
	}

	for qsKey, values := range query {
		if qsKey == "operation" {
			for _, o := range values {
				switch o {
				case "resize":
					var w, h int
					w, _ = queryToInt(query, "width")
					h, _ = queryToInt(query, "height")

					err = trans.Resize(w, h, false)
					if err != nil {
						return trans, err
					}
				case "crop":
					var w, h int
					w, _ = queryToInt(query, "width")
					h, _ = queryToInt(query, "height")

					err = trans.Crop(w, h, query.Get("gravity"), false)
					if err != nil {
						return trans, err
					}
				case "watermark":
					var opacity float64
					opacity, err = strconv.ParseFloat(query.Get("opacity"), 32)
					if err != nil {
						return trans, err
					}
					err = trans.Watermark(query.Get("image"), query.Get("position"), float32(opacity))
					if err != nil {
						return trans, err
					}
				case "blur":
					var sigma, minAmpl float64
					sigma, err = strconv.ParseFloat(query.Get("sigma"), 32)
					if err != nil {
						return trans, err
					}

					minAmpl, _ = strconv.ParseFloat(query.Get("minAmpl"), 32)
					err = trans.Blur(sigma, minAmpl)
					if err != nil {
						return trans, err
					}
				case "rotate":
					var a int
					a, err = queryToInt(query, "angle")
					if err != nil {
						return trans, err
					}
					err = trans.Rotate(a)
					if err != nil {
						return trans, err
					}
				}

			}
			break
		}

	}
	return trans, nil

}
