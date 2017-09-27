package storage

import (
	"io/ioutil"
	"mime"
	"path"
	"regexp"
	"path/filepath"


	"github.com/graymeta/stow"
	fileStorage "github.com/graymeta/stow/local"
	_ "github.com/graymeta/stow/s3"
	Logger "github.com/labstack/gommon/log"

	"mort/object"
	"mort/response"
	"io"
)

var isUrl_RE = regexp.MustCompile("http://")

const notFound = "{\"error\":\"not found\"}"


func Get(obj *object.FileObject) *response.Response {
	key := obj.Key
	//if isUrl_RE.MatchString(key) || obj.UriType != object.URI_TYPE_LOCAL {
	//	return response.NewError(400, errors.New("Not implemented"))
	//}
	//
	//data, err := getFromDisk(obj, key)
	//if os.IsNotExist(err) {
	//	fmt.Println(err)
	//	return response.New(404, []byte(notFound))
	//} else if err != nil {
	//	return response.NewError(503, err)
	//}

	client, err := getClient(obj)
	if err != nil {
		return response.NewError(503, err)
	}

	item, errItem := client.Item(key)
	if errItem != nil {
		Logger.Infof("%s %s %s", errItem, key)
		if errItem == stow.ErrNotFound {
			return response.NewBuf(404, []byte(notFound))
		}

		return response.NewError(544, errItem)
	}

	reader, errOpen := item.Open()
	if errOpen != nil {
		return response.NewError(500, errOpen)
	}

	return prepareResponse(obj, reader)
}

func getClient(obj *object.FileObject)  (stow.Container, error){
	storageCfg := obj.Storage
	var config stow.Config
	var client stow.Location

	if storageCfg.Kind ==  "local" {
		config = stow.ConfigMap{
			fileStorage.ConfigKeyPath: storageCfg.RootPath,
		}
	}

	client, err := stow.Dial(storageCfg.Kind, config)
	if err != nil {
		return nil,  err
	}

	// XXX: check if it is ok
	defer client.Close()

	Logger.Infof("get client %s %s %s", client, obj.Bucket, obj.Key)
	return client.Container(obj.Bucket)
}

func getFromDisk(obj *object.FileObject, filePath string) ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(obj.Storage.RootPath, obj.Bucket,filePath))
}

func prepareResponse(obj *object.FileObject, stream io.Reader) *response.Response {
	res := response.New(200, stream)
	res.SetContentType(mime.TypeByExtension(path.Ext(obj.Key)))
	return res
}
