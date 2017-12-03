package object

import (
	"github.com/aldor007/mort/pkg/config"
	//"github.com/aldor007/mort/pkg/object"
	"net/url"
)

func init() {
	RegisterParser("presets-query", decodePreseQuery)
}

func decodePreseQuery(url *url.URL, mortConfig *config.Config, bucketConfig config.Bucket, obj *FileObject) (bool, error) {
	run, err := decodePreset(url, mortConfig, bucketConfig, obj)
	if run == false || err != nil {
		run, err = decodeQuery(url, mortConfig, bucketConfig, obj)
	}

	return run, err
}
