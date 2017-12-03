package object

import (
	"errors"
	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/log"
	"github.com/aldor007/mort/pkg/transforms"
	//"github.com/aldor007/mort/pkg/object"
	"go.uber.org/zap"
	"net/url"
	"path"
)

func init() {
	RegisterParser("presets", decodePreset)
}

// presetCache cache used presets because we don't need create it always new for each request
var presetCache = make(map[string]transforms.Transforms)

// decodePreset parse given url by matching user defined regexp with request path
func decodePreset(url *url.URL, mortConfig *config.Config, bucketConfig config.Bucket, obj *FileObject) (bool, error) {
	trans := bucketConfig.Transform
	matches := trans.PathRegexp.FindStringSubmatch(obj.Key)
	if matches == nil {
		return false, nil
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
		log.Log().Warn("FileObject decodePreset unknown preset", zap.String("obj.Key", obj.Key), zap.String("parent", parent), zap.String("presetName", presetName),
			zap.String("regexp", trans.Path))
		return true, errors.New("unknown preset " + presetName)
	}

	var err error
	if t, ok := presetCache[presetName]; ok {
		obj.Transforms = t
	} else {
		obj.Transforms, err = presetToTransform(trans.Presets[presetName])
		if err != nil {
			return true, err
		}
		presetCache[presetName] = obj.Transforms
	}

	if trans.ParentBucket != "" {
		parent = "/" + path.Join(trans.ParentBucket, parent)
	} else {
		parent = "/" + parent
	}

	parentObj, err := NewFileObjectFromPath(parent, mortConfig)
	parentObj.Storage = bucketConfig.Storages.Get(trans.ParentStorage)

	if parentObj != nil && bucketConfig.Transform.ResultKey == "hash" {
		obj.Key = hashKey(obj.Transforms.Hash(), subMatchMap["parent"])
	}

	obj.Parent = parentObj
	obj.CheckParent = trans.CheckParent
	return true, err
}

func presetToTransform(preset config.Preset) (transforms.Transforms, error) {
	var trans transforms.Transforms
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
