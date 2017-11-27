package storage

import (
	"encoding/json"
	"github.com/aldor007/stow"
	httpStorage "github.com/aldor007/stow/http"
	fileStorage "github.com/aldor007/stow/local"
	metaStorage "github.com/aldor007/stow/local-meta"
	// import blank to register noop adapter in stow.Register
	_ "github.com/aldor007/stow/noop"
	s3Storage "github.com/aldor007/stow/s3"
	"io"
	"mime"
	"net/http"
	"path"

	"encoding/xml"
	"github.com/aldor007/mort/log"
	"github.com/aldor007/mort/object"
	"github.com/aldor007/mort/response"
	"go.uber.org/zap"
	"strings"
	"time"
)

const notFound = "{\"error\":\"item not found\"}"

// map for used storage client instances
var storageCache = make(map[string]stow.Container)

// Get retrieve obj from given storage and returns its wrapped in response
func Get(obj *object.FileObject) *response.Response {
	key := getKey(obj)
	client, err := getClient(obj)
	if err != nil {
		log.Log().Info("Storage/Get get client", zap.String("obj.Key", obj.Key), zap.String("obj.Bucket", obj.Bucket), zap.Error(err))
		return response.NewError(503, err)
	}

	item, err := client.Item(key)
	if err != nil {
		if err == stow.ErrNotFound {
			log.Log().Info("Storage/Get item response", zap.String("obj.Key", obj.Key), zap.String("obj.Bucket", obj.Bucket), zap.Int("sc", 404))
			return response.NewString(404, notFound)
		}

		log.Log().Info("Storage/Get item response", zap.String("obj.Key", obj.Key), zap.String("obj.Bucket", obj.Bucket), zap.Error(err))
		return response.NewError(500, err)
	}

	reader, err := item.Open()
	if err != nil {
		log.Logs().Warnw("Storage/Get open item", zap.String("obj.Key", obj.Key), zap.String("obj.Bucket", obj.Bucket), zap.Int("sc", 500), zap.Error(err))
		return response.NewError(500, err)
	}

	return prepareResponse(obj, reader, item)
}

// Head retrieve obj from given storage and returns its wrapped in response (but only headers, content of object is omitted)
func Head(obj *object.FileObject) *response.Response {
	key := getKey(obj)
	client, err := getClient(obj)
	if err != nil {
		log.Logs().Infow("Storage/Head get client", zap.String("obj.Key", obj.Key), zap.String("obj.Bucket", obj.Bucket), zap.Error(err))
		return response.NewError(503, err)
	}

	item, err := client.Item(key)
	if err != nil {
		if err == stow.ErrNotFound {
			log.Logs().Infow("Storage/Head item response", zap.String("obj.Key", obj.Key), zap.String("obj.Bucket", obj.Bucket), zap.Int("sc", 404))
			return response.NewString(404, notFound)
		}

		log.Logs().Infow("Storage/Head item response", zap.String("obj.Key", obj.Key), zap.String("obj.Bucket", obj.Bucket), zap.Error(err))
		return response.NewError(500, err)
	}

	return prepareResponse(obj, nil, item)
}

// Set create object on storage wit givent body and headers
func Set(obj *object.FileObject, metaHeaders http.Header, contentLen int64, body io.Reader) *response.Response {
	client, err := getClient(obj)
	if err != nil {
		log.Logs().Warnw("Storage/Set create client", zap.String("obj.Key", obj.Key), zap.String("obj.Bucket", obj.Bucket), zap.Int("sc", 503), zap.Error(err))
		return response.NewError(503, err)
	}

	_, err = client.Put(getKey(obj), body, contentLen, prepareMetadata(obj, metaHeaders))

	if err != nil {
		log.Logs().Warnw("Storage/Set cannot set", zap.String("obj.Key", obj.Key), zap.String("obj.Bucket", obj.Bucket), zap.Int("sc", 500), zap.Error(err))
		return response.NewError(500, err)
	}

	res := response.NewNoContent(200)
	res.SetContentType(metaHeaders.Get("Content-Type"))
	return res
}

// List returns list of object in given path in S3 format
func List(obj *object.FileObject, maxKeys int, delimeter string, prefix string, marker string) *response.Response {
	client, err := getClient(obj)
	if err != nil {
		log.Logs().Warnw("Storage/Set create client", zap.String("obj.Key", obj.Key), zap.String("obj.Bucket", obj.Bucket), zap.Int("sc", 503), zap.Error(err))
		return response.NewError(503, err)
	}

	items, resultMarker, err := client.Items(prefix, marker, maxKeys)
	if err != nil {
		return response.NewError(500, err)
	}

	type contentXML struct {
		Key          string    `xml:"Key"`
		StorageClass string    `xml:"StorageClass"`
		LastModified time.Time `xml:"LastModified"`
		ETag         string    `xml:"ETag"`
		Size         int64     `xml:"Size"`
	}

	type commonPrefixXML struct {
		Prefix string `xml:"Prefix"`
	}

	type listBucketResult struct {
		XMLName        xml.Name          `xml:"ListBucketResult"`
		Name           string            `xml:"Name"`
		Prefix         string            `xml:"Prefix"`
		Marker         string            `xml:"Marker"`
		MaxKeys        int               `xml:"MaxKeys"`
		IsTruncated    bool              `xml:"IsTruncated"`
		Contents       []contentXML      `xml:"Contents"`
		CommonPrefixes []commonPrefixXML `xml:"CommonPrefixes"`
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
			_, ok := commonPrefixes[key]
			if isDir(item) && !ok {
				commonPrefix = key
				commonPrefixes[key] = true
				key = ""
			}
		}

		if key != "" {
			result.Contents = append(result.Contents, contentXML{Key: key, LastModified: lastMod, Size: size, ETag: etag, StorageClass: "STANDARD"})
		}

		if commonPrefix != "" {
			result.CommonPrefixes = append(result.CommonPrefixes, commonPrefixXML{commonPrefix + "/"})
		}

	}

	resultXML, err := xml.Marshal(result)
	if err != nil {
		return response.NewError(500, err)
	}

	res := response.NewBuf(200, resultXML)
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
		allowMetadata := "true"
		config = stow.ConfigMap{
			fileStorage.ConfigKeyPath:      storageCfg.RootPath,
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
		log.Log().Info("Storage/getClient", zap.String("kind", storageCfg.Kind), zap.Error(err))
		return nil, err
	}

	// XXX: check if it is ok
	//defer client.Close()
	bucketName := obj.Bucket
	if storageCfg.Bucket != "" {
		bucketName = storageCfg.Bucket
	}

	container, err := client.Container(bucketName)

	if err != nil {
		log.Log().Info("Storage/getClient error", zap.String("kind", storageCfg.Kind), zap.String("bucket", obj.Bucket), zap.Error(err))
		if err == stow.ErrNotFound && strings.HasPrefix(storageCfg.Kind, "local") {
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
	parseMetadata(obj, metadata, res)

	if err != nil {
		log.Logs().Warnw("Storage/prepareResponse read metadata", zap.String("obj.Key", obj.Key), zap.String("obj.Bucket", obj.Bucket), zap.Int("sc", 500), zap.Error(err))
		return response.NewError(500, err)
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

	if etag != "" {
		res.Set("ETag", etag)
	}
	res.Set("Last-Modified", lastMod.Format(http.TimeFormat))
	res.ContentLength = size

	if contentType, ok := metadata["Content-Type"]; ok {
		res.SetContentType(contentType.(string))
	} else {
		res.SetContentType(mime.TypeByExtension(path.Ext(obj.Key)))
	}

	return res
}

func prepareMetadata(obj *object.FileObject, metaHeaders http.Header) map[string]interface{} {
	metadata := make(map[string]interface{}, len(metaHeaders))
	for k, v := range metaHeaders {
		switch obj.Storage.Kind {
		case "s3":
			keyLower := strings.ToLower(k)
			if strings.HasPrefix(keyLower, "x-amz-meta") || keyLower == "content-type" {
				metadata[strings.Replace(strings.ToLower(k), "x-amz-meta-", "", 1)] = v[0]
			}
		default:
			metadata[k] = v[0]
		}
	}

	return metadata
}

func parseMetadata(obj *object.FileObject, metadata map[string]interface{}, res *response.Response) {
	for k, v := range metadata {
		switch k {
		case "Cache-Control":
			res.Set(k, v.(string))

		}

		if strings.HasPrefix(k, "X-") {
			res.Set(k, v.(string))
		}
	}

	switch obj.Storage.Kind {
	case "s3":
		for k, v := range metadata {
			switch k {
			case "cache-control", "content-type":
				res.Set(k, v.(string))

			}

			res.Set(strings.Join([]string{"x-amz-meta", k}, "-"), v.(string))
		}
	}

}

func isDir(item stow.Item) bool {
	metaData, err := item.Metadata()
	if err != nil {
		return false
	}

	if dir, ok := metaData["is_dir"]; ok {
		return dir.(bool)
	}

	if ct, ok := metaData["content-type"]; ok {
		return ct.(string) == "application/directory"
	}

	size, err := item.Size()
	if err != nil {
		return false
	}

	if size == 0 {
		return true
	}

	return false
}
