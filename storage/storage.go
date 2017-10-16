package storage

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"path"

	"github.com/aldor007/stow"
	fileStorage "github.com/aldor007/stow/local"
	s3Storage "github.com/aldor007/stow/s3"
	httpStorage "mort/storage/http"

	"mort/object"
	"mort/response"
	"mort/log"
	"encoding/xml"
	"time"
)

const notFound = "{\"error\":\"item not found\"}"

func Get(obj *object.FileObject) *response.Response {
	key := getKey(obj)
	client, err := getClient(obj)
	if err != nil {
		log.Log().Infow("Storage/Get get client","obj.Key" ,obj.Key, "error", err)
		return response.NewError(503, err)
	}

	item, err := client.Item(key)
	if err != nil {
		if err == stow.ErrNotFound {
			log.Log().Infow("Storage/Get item response", "obj.Key", obj.Key, "sc", 404)
			return response.NewBuf(404, []byte(notFound))
		}

		log.Log().Infow("Storage/Get item response", "obj.Key", obj.Key, "error", err)
		return response.NewError(500, err)
	}

	metadata, err := item.Metadata()
	if err != nil {
		log.Log().Warnw("Storage/Get read metadata", "obj.Key", obj.Key,"sc", 500, "error", err)
		return response.NewError(500, err)
	}

	reader, err := item.Open()
	if err != nil {
		log.Log().Warnw("Storage/Get open item", "obj.Key", obj.Key, "sc", 500, "error", err)
		return response.NewError(500, err)
	}

	return prepareResponse(obj, reader, metadata)
}

func Set(obj *object.FileObject, _ http.Header, contentLen int64, body io.ReadCloser) *response.Response {
	client, err := getClient(obj)
	if err != nil {
		log.Log().Warnw("Storage/Set create client", "obj.Key", obj.Key, "sc", 503, "error", err)
		return response.NewError(503, err)
	}

	_, err = client.Put(getKey(obj), body, contentLen, nil)

	if err != nil {
		log.Log().Warnw("Storage/Set cannot set" , "obj.Key", obj.Key, "sc", 500, "error", err)
		return response.NewError(500, err)
	}

	res := response.NewBuf(200, []byte(""))
	res.SetContentType(mime.TypeByExtension(path.Ext(obj.Key)))
	return res
}

func List(obj *object.FileObject, maxKeys int, delimeter string, prefix string, marker string) *response.Response {
	client, err := getClient(obj)
	if err != nil {
		log.Log().Warnw("Storage/Set create client", "obj.Key", obj.Key, "sc", 503, "error", err)
		return response.NewError(503, err)
	}

	items, resultMarker, err := client.Items(prefix, "", maxKeys)
	if err != nil {
		return response.NewError(500, err)
	}

	type contentXml struct {
		Key   string `xml:"Key"`
		LastModified time.Time`xml:"LastModified"`
		ETag         string `xml:"ETag"`
		Size         int64 `xml:"Size"`
	}


	type listBucketResult struct {
		XMLName     xml.Name `xml:"ListBucketResult"`
		Name        string   `xml:"Name"`
		Prefix      string   `xml:"Prefix"`
		Marker      string   `xml:"Marker"`
		MaxKeys     int      `xml:"MaxKeys"`
		IsTruncated bool      `xml:"IsTruncated"`
		Contents   []contentXml`xml:"Contents"`
	}

	result := listBucketResult{Name: obj.Bucket, Prefix: prefix, Marker: resultMarker, MaxKeys: maxKeys, IsTruncated: false}


	for _, item := range items {
		lastMod, _ := item.LastMod()
		size, _ := item.Size()
		etag, _ := item.ETag()
		result.Contents = append(result.Contents, contentXml{Key: item.ID(), LastModified: lastMod, Size: size, ETag: etag})
	}

	resultXml, err := xml.Marshal(result)
	if err != nil {
		return response.NewError(500, err)
	}

	res := response.NewBuf(200, resultXml)
	res.SetContentType("application/xml")
	return res
}

func getClient(obj *object.FileObject) (stow.Container, error) {
	storageCfg := obj.Storage
	var config stow.Config
	var client stow.Location

	switch storageCfg.Kind {
	case "local":
		config = stow.ConfigMap{
			fileStorage.ConfigKeyPath: storageCfg.RootPath,
		}
	case "http":
		headers, _ := json.Marshal(storageCfg.Headers)
		config = stow.ConfigMap{
			httpStorage.ConfigUrl:    storageCfg.Url,
			httpStorage.ConfigHeader: string(headers),
		}
	case "s3":
		config = stow.ConfigMap{
			s3Storage.ConfigAccessKeyID: storageCfg.AccessKey,
			s3Storage.ConfigSecretKey:   storageCfg.SecretAccessKey,
			s3Storage.ConfigRegion:      storageCfg.Region,
			s3Storage.ConfigEndpoint:    storageCfg.Endpoint,
		}

	}

	client, err := stow.Dial(storageCfg.Kind, config)
	if err != nil {
		log.Log().Infow("Storage/getClient", "kind", storageCfg.Kind, "error", err)
		return nil, err
	}

	// XXX: check if it is ok
	defer client.Close()

	container, err := client.Container(obj.Bucket)

	if err != nil {
		log.Log().Infow("Storage/getClient ", "kind", storageCfg.Kind, "error", err)
		if err == stow.ErrNotFound && storageCfg.Kind == "local" {
			container, err = client.CreateContainer(obj.Bucket)
			if err != nil {
				return nil, err
			}

			return container,nil
		}

		return  nil, err
	}

	return container, nil
}


func getKey (obj *object.FileObject) string {
	return path.Join(obj.Storage.PathPrefix, obj.Key)
}
func prepareResponse(obj *object.FileObject, stream io.ReadCloser, metadata map[string]interface{}) *response.Response {
	res := response.New(200, stream)

	for k, v := range metadata {
		switch k {
		case  "etag", "last-modified":
			res.Set(k, v.(string))

		}
	}

	if contentType, ok := metadata["content-type"]; ok {
		res.SetContentType(contentType.(string))
	} else {
		res.SetContentType(mime.TypeByExtension(path.Ext(obj.Key)))
	}
	return res
}
