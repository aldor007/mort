package storage

import (
	"io"
	"mime"
	"path"

	"github.com/aldor007/stow"
	fileStorage "github.com/aldor007/stow/local"
	_ "github.com/aldor007/stow/s3"

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

	item, err := client.Item(key)
	if err != nil {
		if err == stow.ErrNotFound {
			return response.NewBuf(404, []byte(notFound))
		}

		return response.NewError(544, err)
	}

	reader, err := item.Open()
	if err != nil {
		return response.NewError(500, err)
	}

	return prepareResponse(obj, reader)
}

func getClient(obj *object.FileObject) (stow.Container, error) {
	storageCfg := obj.Storage
	var config stow.Config
	var client stow.Location

	if storageCfg.Kind == "local" {
		config = stow.ConfigMap{
			fileStorage.ConfigKeyPath: storageCfg.RootPath,
		}
	}

	client, err := stow.Dial(storageCfg.Kind, config)
	if err != nil {
		return nil, err
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
