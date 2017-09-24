package storage

import (
	"fmt"
	"os"
	"regexp"
	"io/ioutil"
	"mime"
	"path"
	"errors"

	Logger "github.com/labstack/gommon/log"
	//"github.com/graymeta/stow"
	//_ "github.com/graymeta/stow/s3"
	//_ "github.com/graymeta/stow/local"

	"mort/object"
	"mort/response"
	"mort/config"
)

var isUrl_RE = regexp.MustCompile("http://")
const notFound = "{\"error\":\"not found\"}"


func Get(obj *object.FileObject) (*response.Response) {
	key := obj.Key
	if isUrl_RE.MatchString(key) || obj.UriType != object.URI_TYPE_LOCAL {
		return response.NewError(400, errors.New("Not implemented"))
	}

	data, err := getFromDisk(key)
	if os.IsNotExist(err) {
		fmt.Println(err)
		return response.New( 404, []byte(notFound))
	} else if err != nil {
		return response.NewError(503, err)
	}

	return prepareResponse(obj, data)
}

func getFromDisk(filePath string) ([]byte, error) {
	return ioutil.ReadFile(config.GetInstance().LocalFilesPath + filePath)
}

func prepareResponse(obj *object.FileObject, data []byte) (*response.Response) {
	res := response.New(200, data)
	res.SetContentType(mime.TypeByExtension(path.Ext(obj.Key)))
	Logger.Infof("Resonse for %s %s", obj.Key, res.Headers)
	return res
}