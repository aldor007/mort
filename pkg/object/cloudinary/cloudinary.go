package cloudinary

import (
	"errors"
	"log"
	"net/url"
	"path"
	"strings"
	"sync"

	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/monitoring"
	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/transforms"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	Decoder struct {
		cache     map[string]transforms.Transforms
		cacheLock sync.RWMutex
	}
)

func init() {
	decoder := newCloudinaryDecoder()
	object.RegisterParser("cloudinary", decoder.decode)
}

func newCloudinaryDecoder() *Decoder {
	return &Decoder{
		cache: make(map[string]transforms.Transforms),
	}
}

func (c *Decoder) getCached(definition string) (transforms.Transforms, bool) {
	c.cacheLock.RLock()
	defer c.cacheLock.RUnlock()
	t, exists := c.cache[definition]
	return t, exists
}

func (c *Decoder) translate(definition string) (transforms.Transforms, error) {
	parser, err := newNotationParser(definition)
	if err != nil {
		return transforms.Transforms{}, err
	}
	transform, err := parser.NextTransform()
	if err != nil {
		return transforms.Transforms{}, err
	}
	if parser.HasNext() {
		return transforms.Transforms{}, errors.New("multiple transforms are not supported")
	}
	return transform, nil
}

// decodePreset parse given url by matching user defined regexp with request path
func (c *Decoder) decode(_ *url.URL, bucketConfig config.Bucket, obj *object.FileObject) (string, error) {
	trans := bucketConfig.Transform
	matches := trans.PathRegexp.FindStringSubmatch(obj.Key)
	if matches == nil {
		return "", nil
	}

	subMatchMap := make(map[string]string)

	for i, name := range trans.PathRegexp.SubexpNames() {
		if i != 0 && name != "" {
			subMatchMap[name] = matches[i]
		}
	}

	transformationsDefinition := subMatchMap["transformations"]

	var err error
	t, ok := c.getCached(transformationsDefinition)
	if ok {
		obj.Transforms = t
	} else {
		obj.Transforms, err = c.translate(transformationsDefinition)
		switch {
		case errors.Is(err, notImplementedError{}):
			monitoring.Log().Error("Cloudinary", append([]zapcore.Field{zap.Error(err)}, obj.LogData()...)...)
		case err == nil, err == errNoToken:
		default:
			monitoring.Log().Info("Cloudinary", append([]zapcore.Field{zap.Error(err)}, obj.LogData()...)...)
			return "", err
		}

		c.cacheLock.Lock()
		c.cache[transformationsDefinition] = obj.Transforms
		c.cacheLock.Unlock()
	}

	parent := subMatchMap["parent"]
	if trans.ParentBucket != "" {
		parent = "/" + path.Join(trans.ParentBucket, parent)
	} else if !strings.HasPrefix(parent, "/") {
		parent = "/" + parent
	}

	log.Printf("tutaj %+v %+v %s %+v\n", matches, subMatchMap, parent, obj.Transforms.NotEmpty)
	return parent, nil
}
