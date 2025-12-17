package object

import (
	"errors"
	"net/url"
	"path"
	"strconv"

	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/transforms"
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
		q, err = queryToInt(query, "quality")
		if err != nil {
			return trans, err
		}
		// Validate quality is in range [1, 100]
		if q < 1 || q > 100 {
			return trans, errors.New("quality must be between 1 and 100")
		}
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
	val := q.Get(k)
	if val == "" {
		return 0, errors.New("empty parameter value for " + k)
	}
	r, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		return 0, err
	}
	return int(r), nil
}

// validatePositiveInt validates that an integer parameter is positive (> 0)
func validatePositiveInt(value int, paramName string) error {
	if value < 0 {
		return errors.New("parameter " + paramName + " cannot be negative")
	}
	if value == 0 {
		return errors.New("parameter " + paramName + " cannot be zero")
	}
	return nil
}

func parseOperation(query url.Values) (transforms.Transforms, error) {
	trans := transforms.New()
	var err error
	opt := query.Get("operation")

	// Check if operation key exists but is empty
	if opValues, hasOperation := query["operation"]; hasOperation && opt == "" {
		// If operation parameter is explicitly provided but empty, return error
		if len(opValues) > 0 && opValues[0] == "" {
			return trans, errors.New("operation parameter cannot be empty")
		}
	}

	if opt == "" {
		var w, h int
		_, hasWidth := query["width"]
		_, hasHeight := query["height"]

		// Parse width if present
		var err1 error
		if hasWidth {
			w, err1 = queryToInt(query, "width")
			if err1 != nil {
				return trans, err1
			}
			if err = validatePositiveInt(w, "width"); err != nil {
				return trans, err
			}
		}

		// Parse height if present
		var err2 error
		if hasHeight {
			h, err2 = queryToInt(query, "height")
			if err2 != nil {
				return trans, err2
			}
			if err = validatePositiveInt(h, "height"); err != nil {
				return trans, err
			}
		}

		// Only apply resize if at least one dimension was provided
		if hasWidth || hasHeight {
			err = trans.Resize(w, h, false, false, false)
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

					// Validate width and height are positive if provided
					if w < 0 || h < 0 {
						return trans, errors.New("width and height cannot be negative")
					}
					if w == 0 && h == 0 {
						return trans, errors.New("at least one of width or height must be specified for resize")
					}

					err = trans.Resize(w, h, false, false, false)
					if err != nil {
						return trans, err
					}
				case "crop":
					var w, h int
					w, _ = queryToInt(query, "width")
					h, _ = queryToInt(query, "height")

					// Validate width and height are positive
					if w < 0 || h < 0 {
						return trans, errors.New("width and height cannot be negative")
					}

					err = trans.Crop(w, h, query.Get("gravity"), false, query.Get("embed") != "")
					if err != nil {
						return trans, err
					}
				case "resizeCropAuto":
					var w, h int
					w, _ = queryToInt(query, "width")
					h, _ = queryToInt(query, "height")

					// Validate width and height are positive
					if w < 0 || h < 0 {
						return trans, errors.New("width and height cannot be negative")
					}

					err = trans.ResizeCropAuto(w, h)
					if err != nil {
						return trans, err
					}
				case "extract":
					var w, h, t, l int
					w, _ = queryToInt(query, "areaWith")
					h, _ = queryToInt(query, "areaHeight")
					t, _ = queryToInt(query, "top")
					l, _ = queryToInt(query, "left")

					// Validate coordinates are non-negative
					if t < 0 || l < 0 || w < 0 || h < 0 {
						return trans, errors.New("extract coordinates cannot be negative")
					}

					err = trans.Extract(t, l, w, h)
					if err != nil {
						return trans, err
					}
				case "watermark":
					var opacity float64
					opacityStr := query.Get("opacity")
					if opacityStr == "" {
						return trans, errors.New("opacity parameter is required for watermark")
					}
					opacity, err = strconv.ParseFloat(opacityStr, 32)
					if err != nil {
						return trans, errors.New("invalid opacity value: " + err.Error())
					}
					// Validate opacity is in range [0, 1]
					if opacity < 0 || opacity > 1 {
						return trans, errors.New("opacity must be between 0 and 1")
					}
					err = trans.Watermark(query.Get("image"), query.Get("position"), float32(opacity))
					if err != nil {
						return trans, err
					}
				case "blur":
					var sigma, minAmpl float64
					sigmaStr := query.Get("sigma")
					if sigmaStr == "" {
						return trans, errors.New("sigma parameter is required for blur")
					}
					sigma, err = strconv.ParseFloat(sigmaStr, 32)
					if err != nil {
						return trans, errors.New("invalid sigma value: " + err.Error())
					}
					// Validate sigma is positive and not too large
					if sigma <= 0 {
						return trans, errors.New("sigma must be positive")
					}
					if sigma > 100 {
						return trans, errors.New("sigma value too large, maximum allowed is 100")
					}

					minAmpl, _ = strconv.ParseFloat(query.Get("minAmpl"), 32)
					err = trans.Blur(sigma, minAmpl)
					if err != nil {
						return trans, err
					}
				case "rotate":
					var a int
					angleStr := query.Get("angle")
					if angleStr == "" {
						return trans, errors.New("angle parameter is required for rotate")
					}
					a, err = queryToInt(query, "angle")
					if err != nil {
						return trans, errors.New("invalid angle value: " + err.Error())
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
