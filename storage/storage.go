package storage

import (
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"os"
	"path"
	"regexp"
	"path/filepath"


	"github.com/graymeta/stow"
	_ "github.com/graymeta/stow/local"
	_ "github.com/graymeta/stow/s3"
	Logger "github.com/labstack/gommon/log"

	"mort/object"
	"mort/response"
)

var isUrl_RE = regexp.MustCompile("http://")

const notFound = "{\"error\":\"not found\"}"

// Dial dials stow storage.
// See stow.Dial for more information.
func Dial(kind string, config stow.Config) (stow.Location, error) {
	return stow.Dial(kind, config)
}

func Get(obj *object.FileObject) *response.Response {
	key := obj.Key
	if isUrl_RE.MatchString(key) || obj.UriType != object.URI_TYPE_LOCAL {
		return response.NewError(400, errors.New("Not implemented"))
	}

	data, err := getFromDisk(obj, key)
	if os.IsNotExist(err) {
		fmt.Println(err)
		return response.New(404, []byte(notFound))
	} else if err != nil {
		return response.NewError(503, err)
	}

	return prepareResponse(obj, data)
}

func getFromDisk(obj *object.FileObject, filePath string) ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(obj.Storage.RootPath, obj.Bucket,filePath))
}

func prepareResponse(obj *object.FileObject, data []byte) *response.Response {
	res := response.New(200, data)
	res.SetContentType(mime.TypeByExtension(path.Ext(obj.Key)))
	Logger.Infof("Resonse for %s %s", obj.Key, res.Headers)
	return res
}
