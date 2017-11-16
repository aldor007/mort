package object

import (
	"errors"
	"path"
	"strings"

	"go.uber.org/zap"
	"mort/config"
	"mort/log"
	"mort/transforms"
	"strconv"
)

func presetToTransform(preset config.PresetsYaml) (transforms.Transforms, error) {
	var trans transforms.Transforms
	filters := preset.Filters

	if len(filters.Thumbnail.Size) > 0 {
		err := trans.Resize(filters.Thumbnail.Size, filters.Thumbnail.Mode == "outbound")
		if err != nil {
			return trans, err
		}
	}

	if len(filters.SmartCrop.Size) > 0 {
		err := trans.Crop(filters.SmartCrop.Size, filters.SmartCrop.Mode == "outbound")
		if err != nil {
			return trans, err
		}
	}

	if len(filters.Crop.Size) > 0 {
		err := trans.Crop(filters.Crop.Size, filters.Crop.Mode == "outbound")
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

	if filters.Blur.Sigma != 0 {
		err := trans.Blur(filters.Blur.Sigma, filters.Blur.MinAmpl)
		if err != nil {
			return trans, err
		}
	}

	if filters.Watermark.Image != "" {
		err := trans.Watermark(filters.Watermark.Image, filters.Watermark.Position, filters.Watermark.Opacity)
		if err != nil {
			return trans, err
		}
	}

	return trans, nil
}

// FileObject is representing parsed request for image or file
//
type FileObject struct {
	Uri         string                `json:"uri"`        // original request path
	Bucket      string                `json:"bucket"`     // request matched bucket
	Key         string                `json:"key"`        // storage path for file
	Transforms  transforms.Transforms `json:"transforms"` // list of transform that should be performed
	Storage     config.Storage        `json:"storage"`    // selected storage that should be used
	Parent      *FileObject           // original image for transformed image
	CheckParent bool                  // boolen if we should always check if parent exists
}

// NewFileObject create new instance of FileObject
// uri should be request path
// mortConfig should be pointer to current buckets config
func NewFileObject(uri string, mortConfig *config.Config) (*FileObject, error) {
	obj := FileObject{}
	obj.Uri = uri
	//obj.uriBytes = []byte(uri)
	obj.CheckParent = false

	err := obj.decode(mortConfig)
	log.Log().Info("FileObject", zap.String("path", uri), zap.String("key", obj.Key), zap.String("bucket", obj.Bucket), zap.String("storage", obj.Storage.Kind),
		zap.Bool("hasTransforms", !obj.Transforms.NotEmpty), zap.Bool("hasParent", obj.HasParent()))
	return &obj, err
}

func (self *FileObject) decode(mortConfig *config.Config) error {
	elements := strings.SplitN(self.Uri, "/", 3)

	self.Bucket = elements[1]
	if len(elements) > 2 {
		self.Key = "/" + elements[2]
	}

	if bucket, ok := mortConfig.Buckets[self.Bucket]; ok {
		err := self.decodeKey(bucket, mortConfig)
		if self.Transforms.NotEmpty {
			self.Storage = bucket.Storages.Transform()
		} else {
			self.Storage = bucket.Storages.Basic()
		}
		return err

	}

	return errors.New("Unknown bucket")
}

func (self *FileObject) decodeKey(bucket config.Bucket, mortConfig *config.Config) error {
	if bucket.Transform == nil {
		return nil
	}

	trans := bucket.Transform
	matches := trans.PathRegexp.FindStringSubmatch(self.Key)
	if matches == nil {
		return nil
	}

	subMatchMap := make(map[string]string, 2)

	for i, name := range trans.PathRegexp.SubexpNames() {
		if i != 0 && name != "" {
			subMatchMap[name] = matches[i]
		}
	}
	presetName := subMatchMap["presetName"] //string(matches[trans.Order.PresetName+1])
	parent := subMatchMap["parent"]         // "/" + string(matches[trans.Order.Parent+1])

	if _, ok := bucket.Transform.Presets[presetName]; !ok {
		log.Log().Warn("FileObject decodeKey unknown preset", zap.String("obj.Key", self.Key), zap.String("parent", parent), zap.String("presetName", presetName),
			zap.String("regexp", trans.Path))
		return errors.New("Unknown preset " + presetName)
	}

	var err error
	self.Transforms, err = presetToTransform(bucket.Transform.Presets[presetName])
	if err != nil {
		return err
	}

	parent = "/" + path.Join(bucket.Transform.ParentBucket, parent)

	parentObj, err := NewFileObject(parent, mortConfig)
	parentObj.Storage = bucket.Storages.Get(bucket.Transform.ParentStorage)

	if parentObj != nil && bucket.Transform.ResultKey == "hash" {
		self.Key = "/" + strings.Join([]string{strconv.FormatUint(uint64(self.Transforms.Hash().Sum64()), 16), subMatchMap["parent"]}, "-")
	}

	self.Parent = parentObj
	self.CheckParent = trans.CheckParent
	return err
}

// HasParent inform if object has parent
func (self *FileObject) HasParent() bool {
	return self.Parent != nil
}

// HasTransform inform if object has transform
func (self *FileObject) HasTransform() bool {
	return self.Transforms.NotEmpty == true
}
