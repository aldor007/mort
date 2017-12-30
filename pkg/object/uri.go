package object

import (
	"bytes"
	"errors"
	"github.com/aldor007/mort/pkg/config"
	"hash"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

var bufPool = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

// ParserFnc is a function that create object from request url
type ParseFnc func(url *url.URL, bucketConfig config.Bucket, obj *FileObject) (string, error)

// parser list of available decoder function
var parsers = make(map[string]ParseFnc)

// Parse pare given url using appropriate parser
// it set object Key, Bucket, Parent and transforms
func Parse(url *url.URL, mortConfig *config.Config, obj *FileObject) error {
	elements := strings.SplitN(url.Path, "/", 3)
	lenElements := len(elements)
	if lenElements < 2 {
		return errors.New("invalid path " + url.Path)
	}

	obj.Bucket = elements[1]
	if lenElements > 2 {
		obj.Key = "/" + elements[2]
		obj.key = elements[2]
	}

	var parent string
	if bucketConfig, ok := mortConfig.Buckets[obj.Bucket]; ok {
		var err error
		var parentObj *FileObject
		if bucketConfig.Transform != nil {
			if fn, ok := parsers[bucketConfig.Transform.Kind]; ok {
				parent, err = fn(url, bucketConfig, obj)
			}

			if err != nil {
				return err
			}

			if parent == "" {
				obj.Storage = bucketConfig.Storages.Basic()
				return err
			}

			parentObj, err = NewFileObjectFromPath(parent, mortConfig)
			parentObj.Storage = bucketConfig.Storages.Get(bucketConfig.Transform.ParentStorage)

			obj.Parent = parentObj
			obj.CheckParent = bucketConfig.Transform.CheckParent
			if obj.Transforms.NotEmpty {
				obj.Storage = bucketConfig.Storages.Transform()
				if obj.allowChangeKey == true && bucketConfig.Transform.ResultKey == "hash" {
					obj.Key = hashKey(obj.Transforms.Hash(), parentObj.key)
				}
			} else {
				obj.Storage = bucketConfig.Storages.Basic()
			}
		} else {
			obj.Storage = bucketConfig.Storages.Basic()
		}

		return err

	}

	return errors.New("unknown bucket")
}

func hashKey(h hash.Hash64, suffix string) string {
	hashB := []byte(strconv.FormatUint(uint64(h.Sum64()), 16))
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	buf.WriteByte('/')
	buf.Write(hashB[0:3])
	buf.WriteByte('/')
	buf.Write(hashB)
	buf.WriteByte('-')
	buf.WriteString(strings.Replace(suffix, "/", "-", -1))
	defer bufPool.Put(buf)
	return buf.String()
}

// RegisterParser add new kind of function to map of decoders and for config validator
func RegisterParser(kind string, fn ParseFnc) {
	parsers[kind] = fn
	config.RegisterTransformKind(kind)
}
