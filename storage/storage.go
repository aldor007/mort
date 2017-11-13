package storage

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"path"
	"github.com/aldor007/stow"
	fileStorage "github.com/aldor007/stow/local"
	metaStorage "github.com/aldor007/stow/local-meta"
	s3Storage "github.com/aldor007/stow/s3"
	httpStorage "mort/storage/http"
	_ "mort/storage/noop"

	"encoding/xml"
	"mort/log"
	"mort/object"
	"mort/response"
	"strings"
	"time"
)

const notFound = "{\"error\":\"item not found\"}"
var storageCache = make(map[string]stow.Container)

func Get(obj *object.FileObject) *response.Response {
	key := getKey(obj)
	client, err := getClient(obj)
	if err != nil {
		log.Log().Infow("Storage/Get get client", "obj.Key", obj.Key, "error", err)
		return response.NewError(503, err)
	}

	item, err := client.Item(key)
	if err != nil {
		if err == stow.ErrNotFound {
			log.Log().Infow("Storage/Get item response", "obj.Key", obj.Key, "sc", 404)
			return response.NewString(404, notFound)
		}

		log.Log().Infow("Storage/Get item response", "obj.Key", obj.Key, "error", err)
		return response.NewError(500, err)
	}

	reader, err := item.Open()
	if err != nil {
		log.Log().Warnw("Storage/Get open item", "obj.Key", obj.Key, "sc", 500, "error", err)
		return response.NewError(500, err)
	}

	return prepareResponse(obj, reader, item)
}

func Head(obj *object.FileObject) *response.Response {
	key := getKey(obj)
	client, err := getClient(obj)
	if err != nil {
		log.Log().Infow("Storage/Get get client", "obj.Key", obj.Key, "error", err)
		return response.NewError(503, err)
	}

	item, err := client.Item(key)
	if err != nil {
		if err == stow.ErrNotFound {
			log.Log().Infow("Storage/Get item response", "obj.Key", obj.Key, "sc", 404)
			return response.NewString(404, notFound)
		}

		log.Log().Infow("Storage/Get item response", "obj.Key", obj.Key, "error", err)
		return response.NewError(500, err)
	}


	return prepareResponse(obj, nil, item)
}

func Set(obj *object.FileObject, headers http.Header, contentLen int64, body io.ReadCloser) *response.Response {
	client, err := getClient(obj)
	if err != nil {
		log.Log().Warnw("Storage/Set create client", "obj.Key", obj.Key, "sc", 503, "error", err)
		return response.NewError(503, err)
	}

	metadata := make(map[string]interface{}, len(headers))
	for k, v := range headers {
		if len(v) == 1 {
			metadata[k] = v[0]
		} else {
			log.Log().Errorf("WTF ---------------- %s %s %s", headers, k, v)
		}
	}

	_, err = client.Put(getKey(obj), body, contentLen, metadata)

	if err != nil {
		log.Log().Warnw("Storage/Set cannot set", "obj.Key", obj.Key, "sc", 500, "error", err)
		return response.NewError(500, err)
	}

	res := response.NewNoContent(200)
	res.SetContentType(mime.TypeByExtension(path.Ext(obj.Key)))
	return res
}

func List(obj *object.FileObject, maxKeys int, delimeter string, prefix string, marker string) *response.Response {
	client, err := getClient(obj)
	if err != nil {
		log.Log().Warnw("Storage/Set create client", "obj.Key", obj.Key, "sc", 503, "error", err)
		return response.NewError(503, err)
	}

	items, resultMarker, err := client.Items(prefix, marker, maxKeys)
	if err != nil {
		return response.NewError(500, err)
	}

	type contentXml struct {
		Key          string    `xml:"Key"`
		StorageClass string    `xml::"StorageClass"`
		LastModified time.Time `xml:"LastModified"`
		ETag         string    `xml:"ETag"`
		Size         int64     `xml:"Size"`
	}

	type commonPrefixXml struct {
		Prefix string `xml:"Prefix"`
	}

	type listBucketResult struct {
		XMLName        xml.Name          `xml:"ListBucketResult"`
		Name           string            `xml:"Name"`
		Prefix         string            `xml:"Prefix"`
		Marker         string            `xml:"Marker"`
		MaxKeys        int               `xml:"MaxKeys"`
		IsTruncated    bool              `xml:"IsTruncated"`
		Contents       []contentXml      `xml:"Contents"`
		CommonPrefixes []commonPrefixXml `xml:"CommonPrefixes"`
	}

	result := listBucketResult{Name: obj.Bucket, Prefix: prefix, Marker: resultMarker, MaxKeys: maxKeys, IsTruncated: false}

	commonPrefixes := make(map[string]bool, len(items))
	for _, item := range items {
		lastMod, _ := item.LastMod()
		size, _ := item.Size()
		etag, _ := item.ETag()
		filePath := strings.Split(item.ID(), "/")
		prefixPath := strings.Split(prefix, "/")
		var commonPrefix string
		var key string


		if len(filePath) > len(prefixPath) {
			key = strings.Join(filePath[0:len(prefixPath)], "/")

			_, ok := commonPrefixes[key]
			if !ok {
				commonPrefix = key
				commonPrefixes[commonPrefix] = true
			} else {
				commonPrefix = ""
			}
			key = ""
		} else {
			key = item.Name()
			// FIXME: add is dir for others adapters
			itemMeta, _ := item.Metadata()
			_, ok := commonPrefixes[key]
			if itemMeta["is_dir"].(bool)  && !ok{
				commonPrefix = key
				commonPrefixes[key] = true
				key = ""
			}
		}

		if key != "" {
			result.Contents = append(result.Contents, contentXml{Key: key, LastModified: lastMod, Size: size, ETag: etag, StorageClass: "STANDARD"})
		}


		if commonPrefix != ""  {
			result.CommonPrefixes = append(result.CommonPrefixes, commonPrefixXml{commonPrefix + "/"})
		}

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
	if c, ok := storageCache[storageCfg.Hash]; ok {
		return c, nil
	}

	var config stow.Config
	var client stow.Location

	switch storageCfg.Kind {
	case "local":
		allowMetadata := ""
		if storageCfg.AllowMetadata {
			allowMetadata = "true"
		}
		config = stow.ConfigMap{
			fileStorage.ConfigKeyPath: storageCfg.RootPath,
			fileStorage.ConfigKeyMetaAllow: allowMetadata,
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
	case "local-meta":
		config = stow.ConfigMap{
			metaStorage.ConfigKeyPath: storageCfg.RootPath,
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
		log.Log().Infow("Storage/getClient error", "kind", storageCfg.Kind, "bucket", obj.Bucket, "error", err)
		if err == stow.ErrNotFound && strings.HasPrefix(storageCfg.Kind, "local")  {
			container, err = client.CreateContainer(obj.Bucket)
			if err != nil {
				return nil, err
			}
			storageCache[storageCfg.Hash] = container
			return container, nil
		}

		return nil, err
	}

	storageCache[storageCfg.Hash] = container
	return container, nil
}

func getKey(obj *object.FileObject) string {
	return path.Join(obj.Storage.PathPrefix, obj.Key)
}

func prepareResponse(obj *object.FileObject, stream io.ReadCloser, item stow.Item) *response.Response {
	res := response.New(200, stream)

	metadata, err := item.Metadata()
	if err != nil {
		log.Log().Warnw("Storage/prepareResponse read metadata", "obj.Key", obj.Key, "sc", 500, "error", err)
		return response.NewError(500, err)
	}

	for k, v := range metadata {
		switch k {
		case "Cache-Control":
			res.Set(k, v.(string))

		}

		if strings.HasPrefix(strings.ToLower(k), "x-") {
			res.Set(k, v.(string))
		}
	}

	etag, err := item.ETag()
	if err != nil {
		return response.NewError(500, err)
	}

	lastMod, err := item.LastMod()
	if err != nil {
		return response.NewError(500, err)
	}

	size, err := item.Size()
	if err != nil {
		return response.NewError(500, err)
	}

	res.Set("ETag", etag)
	res.Set("Last-Modified", lastMod.Format(http.TimeFormat))
	res.ContentLength = size

	if contentType, ok := metadata["Content-Type"]; ok {
		res.SetContentType(contentType.(string))
	} else {
		res.SetContentType(mime.TypeByExtension(path.Ext(obj.Key)))
	}

	return res
}
