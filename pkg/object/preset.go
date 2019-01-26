package object

import (
	"errors"
	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/monitoring"
	"github.com/aldor007/mort/pkg/transforms"
	//"github.com/aldor007/mort/pkg/object"
	"go.uber.org/zap"
	"net/url"
	"path"
	"sync"
	"strings"
)

func init() {
	RegisterParser("presets", decodePreset)
}

// presetCache cache used presets because we don't need create it always new for each request
var presetCache = make(map[string]transforms.Transforms)

// presetCacheLock lock for presetCache
var presetCacheLock = sync.RWMutex{}

// decodePreset parse given url by matching user defined regexp with request path
func decodePreset(_ *url.URL, bucketConfig config.Bucket, obj *FileObject) (string, error) {
	trans := bucketConfig.Transform
	matches := trans.PathRegexp.FindStringSubmatch(obj.Key)
	if matches == nil {
		return "", nil
	}

	subMatchMap := make(map[string]string, 2)

	for i, name := range trans.PathRegexp.SubexpNames() {
		if i != 0 && name != "" {
			subMatchMap[name] = matches[i]
		}
	}

	presetName := subMatchMap["presetName"]
	parent := subMatchMap["parent"]

	if _, ok := trans.Presets[presetName]; !ok {
		monitoring.Log().Warn("FileObject decodePreset unknown preset", zap.String("obj.path", obj.Uri.Path), zap.String("obj.Key", obj.Key), zap.String("parent", parent), zap.String("presetName", presetName),
			zap.String("regexp", trans.Path))
		return "", errors.New("unknown preset " + presetName)
	}

	var err error
	presetCacheLock.RLock()
	if t, ok := presetCache[presetName]; ok {
		obj.Transforms = t
		presetCacheLock.RUnlock()
	} else {
		presetCacheLock.RUnlock()
		obj.Transforms, err = presetToTransform(trans.Presets[presetName])
		if err != nil {
			return parent, err
		}

		presetCacheLock.Lock()
		presetCache[presetName] = obj.Transforms
		presetCacheLock.Unlock()
	}

	if trans.ParentBucket != "" {
		parent = "/" + path.Join(trans.ParentBucket, parent)
	} else if !strings.HasPrefix(parent, "/"){
		parent = "/" + parent
	}

	//if bucketConfig.Transform.ResultKey == "hash" {
	//	obj.Key = hashKey(obj.Transforms.Hash(), subMatchMap["parent"])
	//	obj.allowChangeKey = false
	//}

	return parent, err
}

// presetToTransform convert preset config to transform
// nolint: gocyclo
func presetToTransform(preset config.Preset) (transforms.Transforms, error) {
	trans := transforms.New()
	filters := preset.Filters

	if filters.Thumbnail != nil {
		err := trans.Resize(filters.Thumbnail.Width, filters.Thumbnail.Height, filters.Thumbnail.Mode == "outbound")
		if err != nil {
			return trans, err
		}
	}

	if filters.Crop != nil {
		err := trans.Crop(filters.Crop.Width, filters.Crop.Height, filters.Crop.Gravity, filters.Crop.Mode == "outbound")
		if err != nil {
			return trans, err
		}
	}

	trans.Quality(preset.Quality)

	if filters.Interlace == true {
		err := trans.Interlace()
		if err != nil {
			return trans, err
		}
	}

	if filters.Strip == true {
		err := trans.StripMetadata()
		if err != nil {
			return trans, err
		}
	}

	if preset.Format != "" {
		err := trans.Format(preset.Format)
		if err != nil {
			return trans, err
		}
	}

	if filters.Blur != nil {
		err := trans.Blur(filters.Blur.Sigma, filters.Blur.MinAmpl)
		if err != nil {
			return trans, err
		}
	}

	if filters.Watermark != nil {
		err := trans.Watermark(filters.Watermark.Image, filters.Watermark.Position, filters.Watermark.Opacity)
		if err != nil {
			return trans, err
		}
	}

	if filters.Grayscale {
		trans.Grayscale()
	}

	if filters.Rotate != nil {
		trans.Rotate(filters.Rotate.Angle)
	}

	return trans, nil
}
