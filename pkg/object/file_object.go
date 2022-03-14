package object

import (
	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/monitoring"
	"github.com/aldor007/mort/pkg/transforms"

	//"github.com/aldor007/mort/pkg/uri"
	"context"
	"net/http"
	"net/url"
	"strconv"

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

	monitoring.Log().Info("FileObject", obj.LogData()...)
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
	monitoring.Log().Info("FileObject", obj.LogData()...)
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

//  Type returns type of object "parent" or "transform"
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
}

func (o *FileObject) GetResponseCacheKey() string {
	return o.Bucket + o.Key + o.Range
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
		result = append(result, zap.String("parent.Key", obj.Parent.Key), zap.String("parent.Path", obj.Parent.Uri.Path))
	}

	return append(result, fields...)

}
