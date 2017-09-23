package object

import (
	"strings"
	"regexp"
	"errors"

	"imgserver/config"
	"imgserver/transforms"
	Logger "github.com/labstack/gommon/log"
)

const (
	URI_TYPE_S3 = 0
	URI_TYPE_LOCAL = 1
)

var URI_LIIP_RE = regexp.MustCompile(`\/media\/cache\/.*`)
var URI_LOCAL_RE = regexp.MustCompile(`\/media\/.*`)

func liipToTransform(liip config.LiipFiltersYAML ) ([]transforms.Base, transforms.Param) {
	filters := liip.Filters
	var trans []transforms.Base

	if len(filters.Thumbnail.Size) > 0 {
		trans = append(trans, transforms.Base{Name: "resize", Params: map[string]interface{}{"size": filters.Thumbnail.Size, "mode": filters.Thumbnail.Mode}})
	}

	if len(filters.SmartCrop.Size) > 0 {
		trans = append(trans, transforms.Base{Name: "crop", Params: map[string]interface{}{"size": filters.SmartCrop.Size, "mode": filters.SmartCrop.Mode}})
	}

	if len(filters.Crop.Size) > 0 {
		trans = append(trans, transforms.Base{Name: "crop", Params: map[string]interface{}{"size": filters.Crop.Size, "mode": filters.Crop.Mode}})
	}

	param := transforms.Quailty{Value:liip.Quality}

	return trans, param
}


type FileObject  struct {
	Uri         	string  			`json:"uri"`
	Bucket   		string  			`json:"bucket"`
	Key      		string  			`json:"key"`
	UriType  		int     			`json:"uriType"`
	Parent   		string  			`json:"parent"`
	Transforms 		[]transforms.Base   `json:"transforms"`
	Params 			transforms.Param    `json:"params"`

}

func NewFileObject(path string) (*FileObject, error)  {
	obj := FileObject{}
	obj.Uri = path
	if URI_LOCAL_RE.MatchString(path) {
		obj.UriType = URI_TYPE_LOCAL
	} else {
		obj.UriType = URI_TYPE_S3
	}

	err := obj.decode()
	Logger.Infof("UriType = %d key = %s bucket = %s parent = %s err = %s\n", obj.UriType, obj.Key, obj.Bucket, obj.Parent, err)
	return &obj, err
}

func (self *FileObject) decode() error  {
	if self.UriType == URI_TYPE_LOCAL {
		return self.decodeLiipPath()
	}

	elements := strings.Split(self.Uri, "/")
	if len(elements) < 3 {
		return errors.New("Invalid path")
	}

	self.Bucket = elements[1]
	self.Key = "/" + strings.Join(elements[2:], "/")

	return nil
}

func (self *FileObject) decodeLiipPath() error {
	self.Uri = strings.Replace(self.Uri, "//", "/", 3)
	key := strings.Replace(self.Uri, "/media/cache", "", 1)
	key = strings.Replace(key, "/resolve", "", 1)
	elements := strings.Split(key, "/")
	if URI_LIIP_RE.MatchString(self.Uri) {
		presetName := elements[1]
		//self.Key = strings.Replace(self.Uri, "//", "/", 3)
		self.Key = strings.Replace(self.Uri, "//", "/", 3)
		self.Parent =  "/" + strings.Join(elements[4:], "/")
		liipConfig := config.GetInstance().LiipConfig
		self.Transforms, self.Params = liipToTransform(liipConfig[presetName])
	} else {
		self.Key = self.Uri
	}

	Logger.Debugf("uri: %s parent: %s key: %s len: %d \n", self.Uri, self.Parent, self.Key, len(elements))
	return nil
}

func (self *FileObject) GetParent() *FileObject {
	parent, _ := NewFileObject(self.Parent)
	return parent
}

func (self *FileObject) HasParent() bool{
	return self.Parent != ""
}

