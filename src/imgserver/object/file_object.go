package object

import (
	"strings"
	"regexp"
	"fmt"
	"imgserver/config"
	"imgserver/transforms"
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
		trans = append(trans, transforms.Thumbnail{transforms.ICrop{Size: filters.Thumbnail.Size, Mode: filters.Thumbnail.Mode}})
	}

	if len(filters.SmartCrop.Size) > 0 {
		trans = append(trans, transforms.SmartCrop{transforms.ICrop{Size: filters.SmartCrop.Size, Mode: filters.SmartCrop.Mode}})
	}

	if len(filters.Crop.Size) > 0 {
		trans = append(trans, transforms.Crop{ICrop: transforms.ICrop{Size: filters.Crop.Size, Mode: filters.Crop.Mode}, Start: filters.Crop.Start})
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

func NewFileObject(path string) *FileObject  {
	obj := FileObject{}
	obj.Uri = path
	if URI_LOCAL_RE.MatchString(path) {
		obj.UriType = URI_TYPE_LOCAL
	} else {
		obj.UriType = URI_TYPE_S3
	}
	fmt.Printf("UriType = %d path = %s \n", obj.UriType, path)

	obj.decode()
	return &obj
}

func (self *FileObject) decode() *FileObject  {
	if self.UriType == URI_TYPE_LOCAL {
		return self.decodeLiipPath()
	}
	return self
}

func (self *FileObject) decodeLiipPath() *FileObject {
	self.Uri = strings.Replace(self.Uri, "//", "/", 3)
	key := strings.Replace(self.Uri, "/media/cache", "", 1)
	key = strings.Replace(key, "/resolve", "", 1)
	fmt.Printf("key = %s \n", key)
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
	fmt.Printf("uri: %s parent: %s key: %s len: %d \n", self.Uri, self.Parent, self.Key, len(elements))
	return self
}

func (self *FileObject) GetParent() *FileObject {
	parent := NewFileObject(self.Parent)
	return parent
}

func (self *FileObject) HasParent() bool{
	return self.Parent != ""
}
