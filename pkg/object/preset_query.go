package object

import (
	"github.com/aldor007/mort/pkg/config"
	//"github.com/aldor007/mort/pkg/object"
	"net/url"
)

func init() {
	RegisterParser("presets-query", decodePreseQuery)
}

func decodePreseQuery(url *url.URL, bucketConfig config.Bucket, obj *FileObject) (string, error) {
	parent, err := decodePreset(url, bucketConfig, obj)
	if parent == "" || err != nil {
		parent, err = decodeQuery(url, bucketConfig, obj)
	}

	return parent, err
}
