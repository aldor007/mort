package object

import (
	"errors"
	"github.com/aldor007/mort/config"
	"net/url"
	"strings"
)

// Parse pare given url using appropriate parser
// it set object Key, Bucket, Parent and transforms
func Parse(url *url.URL, mortConfig *config.Config, obj *FileObject) error {
	elements := strings.SplitN(url.Path, "/", 3)

	obj.Bucket = elements[1]
	if len(elements) > 2 {
		obj.Key = "/" + elements[2]
	}

	if bucketConfig, ok := mortConfig.Buckets[obj.Bucket]; ok {
		var err error
		if bucketConfig.Transform != nil {

			switch bucketConfig.Transform.Kind {
			case "presets":
				err = decodePreset(url, mortConfig, bucketConfig, obj)
			case "query":
				err = decodeQuery(url, mortConfig, bucketConfig, obj)
			}
		}

		if obj.Transforms.NotEmpty {
			obj.Storage = bucketConfig.Storages.Transform()
		} else {
			obj.Storage = bucketConfig.Storages.Basic()
		}

		return err

	}

	return errors.New("unknown bucket")
}
