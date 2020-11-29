package object

import (
	"encoding/base64"
	"fmt"
	"github.com/aldor007/mort/pkg/config"
	"path"
	"strings"

	"net/url"
)

func init() {
	RegisterParser("base64_presets", decodeBase64Preset)
}


// decodePreset parse given url by matching user defined regexp with request path
func decodeBase64Preset(u *url.URL, bucketConfig config.Bucket, obj *FileObject) (string, error) {
	trans := bucketConfig.Transform

	matches := trans.PathRegexp.FindStringSubmatch(obj.Key)
	if matches == nil {
		return "", nil
	}

	subMatchMap := make(map[string]string, 2)
	_, err := decodePreset(u, bucketConfig, obj)
	if err != nil {
		return "", err
	}
	for i, name := range trans.PathRegexp.SubexpNames() {
		if i != 0 && name != "" {
			subMatchMap[name] = matches[i]
		}
	}

	parent := subMatchMap["parent"]
	fmt.Println(parent)
	decoded, err := base64.RawStdEncoding.DecodeString(parent)
	if err != nil {
		return "", err
	}
	parent = string(decoded)

	if trans.ParentBucket != "" {
		parent = "/" + path.Join(trans.ParentBucket, parent)
	} else if !strings.HasPrefix(parent, "/") {
		parent = "/" + parent
	}

	return parent, err
}

