package storage

import (
	"io"
	"mime"
	"path"

	"github.com/graymeta/stow"
	fileStorage "github.com/graymeta/stow/local"
	_ "github.com/graymeta/stow/s3"

	"mort/object"
	"mort/response"
)

const notFound = "{\"error\":\"not found\"}"


func Get(obj *object.FileObject) *response.Response {
	key := obj.Key
	client, err := getClient(obj)
	if err != nil {
		return response.NewError(503, err)
	}

	item, errItem := client.Item(key)
	if errItem != nil {
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

	return client.Container(obj.Bucket)
}

func prepareResponse(obj *object.FileObject, stream io.ReadCloser) *response.Response {
	res := response.New(200, stream)
	res.SetContentType(mime.TypeByExtension(path.Ext(obj.Key)))
	return res
}
