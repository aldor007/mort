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
type ParseFnc func(url *url.URL, mortConfig *config.Config, bucketConfig config.Bucket, obj *FileObject) (bool, error)

// parser list of available decoder function
var parsers = make(map[string]ParseFnc)

// Parse pare given url using appropriate parser
// it set object Key, Bucket, Parent and transforms
func Parse(url *url.URL, mortConfig *config.Config, obj *FileObject) error {
	elements := strings.SplitN(url.Path, "/", 3)

	obj.Bucket = elements[1]
	if len(elements) > 2 {
		obj.Key = "/" + elements[2]
		obj.key = elements[2]
	}

	if bucketConfig, ok := mortConfig.Buckets[obj.Bucket]; ok {
		var err error
		if bucketConfig.Transform != nil {
			if fn, ok := parsers[bucketConfig.Transform.Kind]; ok {
				_, err = fn(url, mortConfig, bucketConfig, obj)
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

// RegisterParser add new kind of function to map of decoders
func RegisterParser(kind string, fn ParseFnc) {
	parsers[kind] = fn
}
