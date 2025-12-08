package object

import (
	"math"

	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/monitoring"
	"github.com/aldor007/mort/pkg/transforms"

	//"github.com/aldor007/mort/pkg/uri"
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// FileObject is representing parsed request for image or file
type FileObject struct {
	Uri            *url.URL `json:"uri"`    // original request path
	Bucket         string   `json:"bucket"` // request matched bucket
	Key            string   `json:"key"`    // storage path for file with leading slash
	key            string
	Transforms     transforms.Transforms `json:"transforms"` // list of transform that should be performed
	Storage        config.Storage        `json:"storage"`    // selected storage that should be used
	Parent         *FileObject           // original image for transformed image
	CheckParent    bool                  // boolean if we should always check if parent exists
	AllowChangeKey bool                  // parser can allow or not changing key by this flag
	Debug          bool                  // flag for debug requests
	Ctx            context.Context       // context of request
	Range          string                // HTTP range in request
	RangeData      httpRange             // start, end for HTTP range
}

// NewFileObjectFromPath create new instance of FileObject
// path should be request path
// mortConfig should be pointer to current buckets config
func NewFileObjectFromPath(path string, mortConfig *config.Config) (*FileObject, error) {
	return newFileObjectFromPath(path, mortConfig, true)
}

func NewFileErrorObject(parent string, erroredObject *FileObject) (*FileObject, error) {
	obj := *erroredObject
	obj.key = parent
	obj.Key = parent
	if erroredObject != nil {
		obj.Key += strconv.FormatUint(erroredObject.Transforms.Hash().Sum64(), 16)
		obj.key += strconv.FormatUint(erroredObject.Transforms.Hash().Sum64(), 16)
	}

	return &obj, nil
}
func newFileObjectFromPath(path string, mortConfig *config.Config, allowChangeKey bool) (*FileObject, error) {
	obj := FileObject{}
	obj.Uri = &url.URL{}
	obj.Uri.Path = path

	//obj.uriBytes = []byte(uri)
	obj.CheckParent = false
	obj.AllowChangeKey = allowChangeKey

	err := Parse(obj.Uri, mortConfig, &obj)

	return &obj, err
}

// NewFileObject create new instance of FileObject
// uri is request URL
// mortConfig should be pointer to current buckets config
func NewFileObject(uri *url.URL, mortConfig *config.Config) (*FileObject, error) {
	obj := FileObject{}
	obj.Uri = uri
	//obj.uriBytes = []byte(uri)
	obj.CheckParent = false
	obj.AllowChangeKey = true

	err := Parse(uri, mortConfig, &obj)
	switch {
	case err == errUnknownBucket:
		// continue
	case err != nil:
		monitoring.Log().Error("FileObject", append(obj.LogData(), zap.Error(err))...)
	}

	return &obj, err
}

// HasParent inform if object has parent
func (o *FileObject) HasParent() bool {
	return o.Parent != nil
}

// HasTransform inform if object has transform
func (o *FileObject) HasTransform() bool {
	return o.Transforms.NotEmpty
}

// Type returns type of object "parent" or "transform"
func (o *FileObject) Type() string {
	if o.HasTransform() {
		return "transform"
	}
	return "parent"
}

// AppendToKey add string to key
func (o *FileObject) AppendToKey(str string) {
	o.key = o.key + str
	o.Key = o.Key + str
}

// FillWithRequest assign to object request and HTTP range data
func (o *FileObject) FillWithRequest(req *http.Request, ctx context.Context) {
	o.Ctx = ctx
	o.Range = req.Header.Get("Range")
	if o.Range != "" {
		var err error
		o.RangeData, err = parseRange(o.Range)
		if err != nil {
			monitoring.Log().Error("FileObject unable to parse range", append(o.LogData(), zap.Error(err))...)
		}
	}

	parent := o.Parent
	for parent != nil {
		parent.Ctx = ctx
		parent = parent.Parent
	}

}

func (o *FileObject) GetResponseCacheKey() string {
	// Use strings.Builder to avoid multiple allocations from string concatenation
	var b strings.Builder
	b.Grow(len(o.Bucket) + len(o.Key) + len(o.Range))
	b.WriteString(o.Bucket)
	b.WriteString(o.Key)
	b.WriteString(o.Range)
	return b.String()
}

func (o *FileObject) Copy() *FileObject {
	copy := FileObject{
		Uri:            o.Uri,
		Bucket:         o.Bucket,
		Key:            o.Key,
		key:            o.key,
		Transforms:     o.Transforms,
		Storage:        o.Storage,
		Parent:         o.Parent,
		CheckParent:    o.CheckParent,
		AllowChangeKey: o.AllowChangeKey,
		Debug:          o.Debug,
		Ctx:            context.Background(),
		Range:          o.Range,
	}

	return &copy
}

// LogData log data for given object
func (obj *FileObject) LogData(fields ...zapcore.Field) []zapcore.Field {
	result := []zapcore.Field{zap.String("obj.path", obj.Uri.Path), zap.String("obj.Key", obj.Key), zap.String("obj.Bucket", obj.Bucket), zap.String("obj.Storage", obj.Storage.Kind),
		zap.Bool("obj.HasTransforms", obj.HasTransform()), zap.Bool("obj.HasParent", obj.HasParent())}

	if obj.HasParent() {
		result = append(result, zap.String("parent.Key", obj.Parent.Key), zap.String("parent.Path", obj.Parent.Uri.Path), zap.String("parent.Storage", obj.Parent.Storage.Kind))
	}

	if obj.Range != "" {
		result = append(result, zap.String("obj.Range", obj.Range))
	}

	return append(result, fields...)

}

type httpRange struct {
	Start uint64
	End   uint64
}

func parseRange(s string) (httpRange, error) {
	var httpRangeData httpRange
	if s == "" {
		return httpRangeData, nil // header not present
	}

	if !strings.HasPrefix(s, "bytes=") {
		return httpRangeData, errors.New("invalid range")
	}
	if strings.Index(s, "-") == -1 {
		return httpRangeData, errors.New("invalid range")
	}
	var minStart uint64
	minStart = math.MaxUint64
	maxEnd := uint64(0)

	// Parse ranges more efficiently without extra allocations
	rangesStr := s[6:] // Skip "bytes="
	for len(rangesStr) > 0 {
		// Find next range (comma-separated)
		commaIdx := strings.IndexByte(rangesStr, ',')
		var ra string
		if commaIdx >= 0 {
			ra = rangesStr[:commaIdx]
			rangesStr = rangesStr[commaIdx+1:]
		} else {
			ra = rangesStr
			rangesStr = ""
		}

		// Trim spaces without allocating
		ra = strings.TrimSpace(ra)
		if ra == "" {
			continue
		}

		i := strings.IndexByte(ra, '-')
		if i < 0 {
			return httpRangeData, errors.New("invalid range")
		}

		// Parse start and end without TrimSpace allocations
		startStr := strings.TrimSpace(ra[:i])
		endStr := strings.TrimSpace(ra[i+1:])

		if startStr != "" {
			// If no start is specified, end specifies the
			// range start relative to the end of the file.
			startVal, err := strconv.ParseUint(startStr, 10, 64)
			if err != nil {
				return httpRangeData, errors.New("invalid range")
			}
			if startVal < minStart {
				minStart = startVal
			}

			if startVal > maxEnd {
				maxEnd = startVal
			}
		}
		if endStr != "" {
			// If no start is specified, end specifies the
			// range start relative to the end of the file.
			endVal, err := strconv.ParseUint(endStr, 10, 64)
			if err != nil {
				return httpRangeData, errors.New("invalid range")
			}

			if endVal > maxEnd {
				maxEnd = endVal
			}
		}
	}
	httpRangeData.Start = uint64(minStart)
	httpRangeData.End = uint64(maxEnd)

	return httpRangeData, nil
}
