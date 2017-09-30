package object

import (
	"errors"
	"strings"

	Logger "github.com/labstack/gommon/log"
	"mort/config"
	"mort/transforms"
)

func presetToTransform(preset config.PresetsYaml) transforms.Transforms {
	var trans transforms.Transforms
	filters := preset.Filters

	if len(filters.Thumbnail.Size) > 0 {
		trans.ResizeT(filters.Thumbnail.Size, filters.Thumbnail.Mode == "outbound")
	}

	if len(filters.SmartCrop.Size) > 0 {
		trans.CropT(filters.SmartCrop.Size, filters.SmartCrop.Mode == "outbound")
	}

	if len(filters.Crop.Size) > 0 {
		trans.CropT(filters.Crop.Size, filters.Crop.Mode == "outbound")
	}

	trans.Quality = preset.Quality

	return trans
}

type FileObject struct {
	Uri        string                `json:"uri"`
	Bucket     string                `json:"bucket"`
	Key        string                `json:"key"`
	Transforms transforms.Transforms `json:"transforms"`
	Storage    config.Storage        `json:"storage"`
	Parent     *FileObject
}

func NewFileObject(path string, mortConfig *config.Config) (*FileObject, error) {
	obj := FileObject{}
	obj.Uri = path

	err := obj.decode(mortConfig)
	Logger.Infof("key = %s bucket = %s parent = %s\n", obj.Key, obj.Bucket, obj.Parent)
	return &obj, err
}

func (self *FileObject) decode(mortConfig *config.Config) error {
	elements := strings.Split(self.Uri, "/")
	if len(elements) < 3 {
		return errors.New("Invalid path")
	}

	self.Bucket = elements[1]
	self.Key = "/" + strings.Join(elements[2:], "/")
	if bucket, ok := mortConfig.Buckets[self.Bucket]; ok {
		self.decodeKey(bucket, mortConfig)
		if self.HasTransform() {
			self.Storage = bucket.Storages.Transform
		} else {
			self.Storage = bucket.Storages.Basic
		}

	} else {
		return errors.New("Unknown bucket")
	}

	return nil
}

func (self *FileObject) decodeKey(bucket config.Bucket, mortConfig *config.Config) error {
	if bucket.Transform == nil {
		return nil
	}

	trans := bucket.Transform
	matches := trans.PathRegexp.FindStringSubmatch(self.Key)
	if len(matches) < 3 {
		return nil
	}

	presetName := string(matches[trans.Order.PresetName+1])
	parent := "/" + string(matches[trans.Order.Parent+1])

	self.Transforms = presetToTransform(bucket.Transform.Presets[presetName])
	parentObj, _ := NewFileObject(parent, mortConfig)
	self.Parent = parentObj
	return nil
}

func (self *FileObject) HasParent() bool {
	return self.Parent != nil
}

func (self *FileObject) HasTransform() bool {
	return self.Transforms.NotEmpty == true
}
