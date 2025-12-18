package storage

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/aldor007/stow"
	azureStorage "github.com/aldor007/stow/azure"
	googleStorage "github.com/aldor007/stow/google"
	httpStorage "github.com/aldor007/stow/http"
	fileStorage "github.com/aldor007/stow/local"
	metaStorage "github.com/aldor007/stow/local-meta"
	oracleStorage "github.com/aldor007/stow/oracle"
	sftpStorage "github.com/aldor007/stow/sftp"

	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/glacier"
	"github.com/aldor007/mort/pkg/monitoring"
	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/response"
	b2Storage "github.com/aldor007/stow/b2"
	_ "github.com/aldor007/stow/noop"
	s3Storage "github.com/aldor007/stow/s3"
	"go.uber.org/zap"
)

const notFound = "{\"error\":\"item not found\"}"

// storageClient struct that contain location and container
type storageClient struct {
	container stow.Container
	client    stow.Location
}

type responseData struct {
	statusCode int
	stream     io.ReadCloser
	item       stow.Item
	headers    http.Header
}

func newResponseData() responseData {
	r := responseData{}
	r.statusCode = 200
	r.headers = make(http.Header)
	return r
}

// storageCache map for used storage client instances
var storageCache sync.Map // map[string]*storageClientEntry

// storageClientEntry wraps a storage client with initialization state
type storageClientEntry struct {
	once   sync.Once
	client storageClient
	err    error
}

// handleGlacierError detects and handles GLACIER/archive storage class errors
// Returns 503 with Retry-After header and initiates restore if configured
func handleGlacierError(obj *object.FileObject, err error, item stow.Item) *response.Response {
	// Check if this is a GLACIER error
	if !strings.Contains(err.Error(), "InvalidObjectState") {
		return nil // Not a GLACIER error
	}

	monitoring.Report().Inc("glacier_error_detected")

	// Get bucket configuration
	mortConfig := config.GetInstance()
	bucket, ok := mortConfig.Buckets[obj.Bucket]
	if !ok || bucket.Glacier == nil || !bucket.Glacier.Enabled {
		// No GLACIER config or disabled - return generic error
		return response.NewError(503, fmt.Errorf("object in GLACIER storage class"))
	}

	glacierCfg := bucket.Glacier

	// Check if restore already in progress via cache
	cache := glacier.GetRestoreCache(mortConfig.Server.Cache)
	status, _ := cache.GetRestoreStatus(obj.Ctx, obj.Key)

	if status == nil || !status.InProgress {
		// Initiate restore using stow Restorable interface
		if restorable, ok := item.(stow.Restorable); ok {
			go func() {
				// Use background context with timeout for restore request
				restoreCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				monitoring.Log().Info("Initiating GLACIER restore",
					obj.LogData(
						zap.String("tier", glacierCfg.RestoreTier),
						zap.Int("days", glacierCfg.RestoreDays),
					)...)

				if err := restorable.Restore(restoreCtx, glacierCfg.RestoreDays, glacierCfg.RestoreTier); err != nil {
					monitoring.Log().Error("GLACIER restore failed", obj.LogData(zap.Error(err))...)
					return
				}

				// Mark restore as requested in cache
				expiration := time.Duration(glacierCfg.RetryAfterSeconds) * time.Second
				if err := cache.MarkRestoreRequested(restoreCtx, obj.Key, expiration); err != nil {
					monitoring.Log().Warn("Failed to cache restore status", obj.LogData(zap.Error(err))...)
				}

				monitoring.Report().Inc("glacier_restore_initiated")
			}()
		} else {
			monitoring.Log().Warn("Item does not implement Restorable interface", obj.LogData()...)
		}
	} else {
		monitoring.Log().Info("GLACIER restore already in progress (cached)",
			obj.LogData(
				zap.Time("requestedAt", status.RequestedAt),
				zap.Time("expiresAt", status.ExpiresAt),
			)...)
	}

	// Return 503 with Retry-After header
	res := response.NewError(503, fmt.Errorf("object in GLACIER storage class, restore in progress"))
	res.Set("Retry-After", fmt.Sprintf("%d", glacierCfg.RetryAfterSeconds))
	res.Set("X-Mort-Glacier-Status", "restoring")
	res.Set("X-Mort-Glacier-Tier", glacierCfg.RestoreTier)

	return res
}

// Get retrieve obj from given storage and returns its wrapped in response
func Get(obj *object.FileObject) *response.Response {
	inc(obj, "get")
	metric := "storage_time;method:get,storage:" + obj.Storage.Kind
	t := monitoring.Report().Timer(metric)
	defer t.Done()
	key := getKey(obj)
	instance, err := getClient(obj)
	client := instance.container
	if err != nil {
		monitoring.Log().Info("Storage/Get get client", obj.LogData(zap.Error(err))...)
		return response.NewError(503, fmt.Errorf("unable to get client %s, err: %v", obj.Key, err))
	}

	item, err := client.Item(key)
	if err != nil {
		if err == stow.ErrNotFound {
			monitoring.Log().Info("Storage/Get item response", zap.String("obj.Storage.Kind", obj.Storage.Kind), zap.String("obj.Key", obj.Key), zap.String("key", key), zap.String("obj.Bucket", obj.Bucket), zap.Int("statusCode", 404))
			return response.NewString(404, notFound)
		}

		monitoring.Log().Info("Storage/Get item response", obj.LogData(zap.Error(err))...)
		return response.NewError(500, fmt.Errorf("get item %s, err %v", obj.Key, err))
	}

	if isDir(item) {
		res := response.NewNoContent(404)
		res.SetContentType("application/xml")
		return res
	}

	resData := newResponseData()
	resData.item = item
	var responseStream io.ReadCloser
	if instance.client.HasRanges() && obj.Range != "" {
		var stowRanger stow.ItemRanger
		stowRanger = item.(stow.ItemRanger)
		responseStream, err = stowRanger.OpenRange(obj.RangeData.Start, obj.RangeData.End)
		resData.statusCode = 206
	} else {
		responseStream, err = item.Open()
		resData.statusCode = 200
	}
	if err != nil {
		if responseStream != nil {
			responseStream.Close()
		}

		// Check if this is a GLACIER error and handle if configured
		if obj.Storage.Kind == "s3" {
			if glacierRes := handleGlacierError(obj, err, item); glacierRes != nil {
				return glacierRes
			}
		}

		monitoring.Log().Warn("Storage/Get open item", obj.LogData(zap.Int("statusCode", 500), zap.Error(err))...)
		return response.NewError(500, fmt.Errorf("unable to open item %s err: %v", obj.Key, err))
	}
	resData.stream = responseStream
	return prepareResponse(obj, resData)
}

// Head retrieve obj from given storage and returns its wrapped in response (but only headers, content of object is omitted)
func Head(obj *object.FileObject) *response.Response {
	inc(obj, "head")
	metric := "storage_time;method:head,storage:" + obj.Storage.Kind
	t := monitoring.Report().Timer(metric)
	defer t.Done()
	key := getKey(obj)
	instance, err := getClient(obj)
	client := instance.container
	if err != nil {
		monitoring.Log().Info("Storage/Head get client", obj.LogData(zap.Error(err))...)
		return response.NewError(503, err)
	}

	item, err := client.Item(key)
	if err != nil {
		if err == stow.ErrNotFound {
			monitoring.Log().Info("Storage/Head item response", obj.LogData(zap.String("key", key), zap.Int("statusCode", 404))...)
			return response.NewString(404, notFound)
		}

		monitoring.Log().Info("Storage/Head item response", obj.LogData(zap.Error(err))...)
		return response.NewError(500, err)
	}
	resData := newResponseData()
	resData.item = item
	resData.statusCode = 200
	return prepareResponse(obj, resData)
}

// Set create object on storage wit given body and headers
func Set(obj *object.FileObject, metaHeaders http.Header, contentLen int64, body io.Reader) *response.Response {
	inc(obj, "set")
	metric := "storage_time;method:set,storage:" + obj.Storage.Kind
	t := monitoring.Report().Timer(metric)
	defer t.Done()
	monitoring.Report().Gauge("storage_throughput;method:set,storage:"+obj.Storage.Kind, float64(contentLen))
	instance, err := getClient(obj)
	client := instance.container
	if err != nil {
		monitoring.Log().Warn("Storage/Set create client", obj.LogData(zap.Int("statusCode", 503), zap.Error(err))...)
		return response.NewError(503, err)
	}

	key := getKey(obj)
	switch obj.Storage.Kind {
	case "s3":
		// in such case we want to create dir but s3 is key/value store so it is not handling it
		if contentLen == 0 && strings.HasSuffix(key, "/") {
			res := response.NewNoContent(200)
			return res
		}

	}

	if len(obj.Storage.Headers) != 0 {
		for k, v := range obj.Storage.Headers {
			metaHeaders.Set(k, v)
		}
	}
	_, err = client.Put(getKey(obj), body, contentLen, prepareMetadata(obj, metaHeaders))

	if err != nil {
		monitoring.Log().Warn("Storage/Set cannot set", obj.LogData(zap.Int("statusCode", 500), zap.Error(err))...)
		return response.NewError(500, err)
	}

	res := response.NewNoContent(200)
	res.SetContentType(metaHeaders.Get("Content-Type"))
	return res
}

// Delete remove object from given storage
func Delete(obj *object.FileObject) *response.Response {
	inc(obj, "delete")
	metric := "storage_time;method:delete,storage:" + obj.Storage.Kind
	t := monitoring.Report().Timer(metric)
	defer t.Done()
	instance, err := getClient(obj)
	client := instance.container
	if err != nil {
		monitoring.Log().Warn("Storage/Delete create client", obj.LogData(zap.Int("statusCode", 503), zap.Error(err))...)
		return response.NewError(503, err)
	}

	resHead := Head(obj)
	if resHead.StatusCode == 200 {
		err = client.RemoveItem(getKey(obj))

		if err != nil {
			monitoring.Log().Warn("Storage/Delete cannot delete", obj.LogData(zap.Int("statusCode", 500), zap.Error(err))...)
			return response.NewError(500, err)
		}
	} else if resHead.StatusCode == 404 {
		res := response.NewNoContent(200)
		return res
	}

	return resHead
}

func CreatePreSign(obj *object.FileObject) *response.Response {
	inc(obj, "presign")
	metric := "storage_time;method:presign,storage:" + obj.Storage.Kind
	t := monitoring.Report().Timer(metric)
	defer t.Done()
	instance, err := getClient(obj)
	client := instance.container
	if err != nil {
		monitoring.Log().Warn("Storage/CreatePresign create client", obj.LogData(zap.Int("statusCode", 503), zap.Error(err))...)
		return response.NewError(503, err)
	}

	uri, err := client.PreSignRequest(obj.Ctx, stow.ClientMethodGet, getKey(obj), stow.PresignRequestParams{
		ExpiresIn: time.Hour * 5,
	})
	if err != nil {
		monitoring.Log().Warn("Storage/CreatePresign create request", obj.LogData(zap.Int("statusCode", 503), zap.Error(err))...)
		return response.NewError(503, err)
	}

	res := response.NewNoContent(307)
	res.Headers.Set("location", uri)

	return res
}

// List returns list of object in given path in S3 format
// nolint: gocyclo
func List(obj *object.FileObject, maxKeys int, _ string, prefix string, marker string) *response.Response {
	instance, err := getClient(obj)
	client := instance.container
	if err != nil {
		monitoring.Log().Warn("Storage/List", obj.LogData(zap.Int("statusCode", 503), zap.Error(err))...)
		return response.NewError(503, err)
	}

	prefix = path.Join(obj.Storage.PathPrefix, prefix)

	if prefix != "" && prefix != "/" && obj.Storage.Kind == "local-meta" {
		_, err = client.Item(prefix)
		if err != nil {
			if err == stow.ErrNotFound {
				monitoring.Log().Info("Storage/List item not fountresponse", obj.LogData(zap.Int("statusCode", 404))...)
				return response.NewString(404, obj.Key)
			}
		}
	}

	items, resultMarker, err := client.Items(prefix, marker, maxKeys)
	if err != nil {
		monitoring.Log().Warn("Storage/List", obj.LogData(zap.Int("statusCode", 500), zap.Error(err))...)
		return response.NewError(500, err)
	}

	type contentXML struct {
		Key          string `xml:"Key"`
		StorageClass string `xml:"StorageClass"`
		LastModified string `xml:"LastModified"`
		ETag         string `xml:"ETag"`
		Size         int64  `xml:"Size"`
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
	// Preallocate result slices with estimated capacity
	result.Contents = make([]contentXML, 0, len(items))
	result.CommonPrefixes = make([]commonPrefixXML, 0, len(items)/2)

	// Split prefix once before loop instead of repeating for every item
	prefixPath := strings.Split(prefix, "/")

	for _, item := range items {
		lastMod, _ := item.LastMod()
		size, _ := item.Size()
		etag, _ := item.ETag()
		itemID := item.ID()
		filePath := strings.Split(itemID, "/")
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
		} else {
			key = item.Name()
			_, ok := commonPrefixes[key]
			if isDir(item) && !ok {
				commonPrefix = key
				commonPrefixes[key] = true
				//key = key + "/"
			}
		}

		if itemID[len(itemID)-1] == '/' {
			key = key + "/"
			size = 0
		}

		if key != "" {
			result.Contents = append(result.Contents, contentXML{Key: key, LastModified: lastMod.Format(time.RFC3339), Size: size, ETag: etag, StorageClass: "STANDARD"})
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

func getClient(obj *object.FileObject) (storageClient, error) {
	storageCfg := obj.Storage

	// Load or create entry atomically
	entryInterface, _ := storageCache.LoadOrStore(storageCfg.Hash, &storageClientEntry{})
	entry := entryInterface.(*storageClientEntry)

	// Initialize client exactly once per storage config
	entry.once.Do(func() {
		entry.client, entry.err = createStorageClient(obj, storageCfg)
	})

	return entry.client, entry.err
}

func createStorageClient(obj *object.FileObject, storageCfg config.Storage) (storageClient, error) {
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
			s3Storage.ConfigHTTPTracing: storageCfg.HTTPTracing,
		}
	case "local-meta":
		config = stow.ConfigMap{
			metaStorage.ConfigKeyPath: storageCfg.RootPath,
		}
	case "b2":
		config = stow.ConfigMap{
			b2Storage.ConfigAccountID:      storageCfg.B2AccountID,
			b2Storage.ConfigApplicationKey: storageCfg.B2ApplicationKey,
			b2Storage.ConfigKeyID:          storageCfg.B2ApplicationKeyID,
		}
	case "google":
		config = stow.ConfigMap{
			googleStorage.ConfigJSON:      storageCfg.GoogleConfigJSON,
			googleStorage.ConfigProjectId: storageCfg.GoogleProjectID,
			googleStorage.ConfigScopes:    storageCfg.GoogleScopes,
		}
	case "oracle":
		config = stow.ConfigMap{
			oracleStorage.ConfigUsername:     storageCfg.OracleUsername,
			oracleStorage.ConfigPassword:     storageCfg.OraclePassword,
			oracleStorage.ConfigAuthEndpoint: storageCfg.OracleAuthEndpoint,
		}
	case "sftp":
		config = stow.ConfigMap{
			sftpStorage.ConfigHost:                 storageCfg.SFTPHost,
			sftpStorage.ConfigPort:                 storageCfg.SFTPPort,
			sftpStorage.ConfigUsername:             storageCfg.SFTPUsername,
			sftpStorage.ConfigPassword:             storageCfg.SFTPPassword,
			sftpStorage.ConfigPrivateKey:           storageCfg.SFTPPrivateKey,
			sftpStorage.ConfigPrivateKeyPassphrase: storageCfg.SFTPPrivateKeyPass,
			sftpStorage.ConfigHostPublicKey:        storageCfg.SFTPHostPublicKey,
			sftpStorage.ConfigBasePath:             storageCfg.SFTPHostBasePath,
		}
	case "azure":
		config = stow.ConfigMap{
			azureStorage.ConfigAccount: storageCfg.AzureAccount,
			azureStorage.ConfigKey:     storageCfg.AzureKey,
		}

	}

	client, err := stow.Dial(storageCfg.Kind, config)
	if err != nil {
		monitoring.Log().Info("Storage/getClient", zap.String("kind", storageCfg.Kind), zap.Error(err))
		return storageClient{}, err
	}

	// XXX: check if it is ok
	//defer client.Close()
	bucketName := obj.Bucket
	if storageCfg.Bucket != "" {
		bucketName = storageCfg.Bucket
	}

	container, err := client.Container(bucketName)

	if err != nil {
		monitoring.Log().Info("Storage/getClient container get error", zap.String("kind", storageCfg.Kind), zap.String("bucket", bucketName), zap.Error(err))
		if err == stow.ErrNotFound && strings.HasPrefix(storageCfg.Kind, "local") {
			container, err = client.CreateContainer(obj.Bucket)
			if err != nil {
				return storageClient{}, err
			}
			storageInstance := storageClient{container, client}
			return storageInstance, nil
		}

		return storageClient{}, err
	}

	storageInstance := storageClient{container, client}
	return storageInstance, nil
}

func getKey(obj *object.FileObject) string {
	switch obj.Storage.Kind {
	case "b2", "s3":
		return strings.TrimPrefix(path.Join(obj.Storage.PathPrefix, obj.Key), "/")
	default:
		return path.Join(obj.Storage.PathPrefix, obj.Key)

	}
}

func prepareResponse(obj *object.FileObject, resData responseData) *response.Response {
	res := response.New(resData.statusCode, resData.stream)

	item := resData.item
	metadata, err := item.Metadata()

	if err != nil {
		monitoring.Log().Warn("Storage/prepareResponse read metadata error", obj.LogData(zap.Int("statusCode", 500), zap.Error(err))...)
		return response.NewError(500, fmt.Errorf("metadata read err %v", err))
	}

	parseMetadata(obj, metadata, res)

	etag, err := item.ETag()
	if err != nil {
		monitoring.Log().Warn("Storage/prepareResponse read etag error", obj.LogData(zap.Int("statusCode", 500), zap.Error(err))...)
		return response.NewError(500, err)
	}

	lastMod, err := item.LastMod()
	if err != nil {
		monitoring.Log().Warn("Storage/prepareResponse read lastmod error", obj.LogData(zap.Int("statusCode", 500), zap.Error(err))...)
		return response.NewError(500, err)
	}

	if resData.statusCode == http.StatusPartialContent {
		contentRange, err := stow.GetContentRange(item, obj.RangeData.Start, obj.RangeData.End)
		if err != nil {
			monitoring.Log().Warn("Storage/prepareResponse read content range data error fallback to normal response", obj.LogData(zap.Int("statusCode", 500), zap.Error(err))...)
			res.StatusCode = http.StatusOK
		} else {
			res.Set("content-range", contentRange.ContentRange)
			res.ContentLength = contentRange.ContentLength
		}
	} else {
		size, err := item.Size()
		if err == nil {
			res.ContentLength = size
		}
	}

	var resSize int64

	if res.ContentLength != 0 && res.ContentLength != -1 {
		resSize = res.ContentLength
	} else {
		resSize, _ = item.Size()
	}

	if resData.stream != nil {
		monitoring.Report().Gauge("storage_throughput;method:get,storage:"+obj.Storage.Kind, float64(resSize))
	}

	if etag != "" {
		res.Set("ETag", etag)
	}
	res.Set("Last-Modified", lastMod.UTC().Format(http.TimeFormat))

	if contentType, ok := metadata["Content-Type"]; ok {
		res.SetContentType(contentType.(string))
	} else if contentType, ok := metadata["content-type"]; ok {
		res.SetContentType(contentType.(string))
	} else {
		ct := mime.TypeByExtension(path.Ext(obj.Uri.Path))
		if ct != "" {
			res.SetContentType(ct)
		} else {
			if isDir(item) {
				res.SetContentType("application/directory")
			}
		}
	}
	return res
}

func prepareMetadata(obj *object.FileObject, metaHeaders http.Header) map[string]interface{} {
	metadata := make(map[string]interface{}, len(metaHeaders))
	for k, v := range metaHeaders {
		switch obj.Storage.Kind {
		case "s3":
			keyLower := strings.ToLower(k)
			if keyLower == "content-type" || keyLower == "content-md5" || keyLower == "content-disposition" {
				metadata[keyLower] = v[0]
			} else if strings.HasPrefix(keyLower, "x-amz-meta") {
				metadata[strings.Replace(keyLower, "x-amz-meta-", "", 1)] = v[0]
			} else if strings.HasPrefix(keyLower, "x-amz") {
				switch keyLower {
				case "x-amz-date", "x-amz-content-sha256":
				default:
					metadata[keyLower] = v[0]
				}
			}
		default:
			keyLower := strings.ToLower(k)
			if strings.HasPrefix(keyLower, "x-amz-meta") || keyLower == "content-type" || keyLower == "etag" {
				metadata[keyLower] = v[0]
			}
		}
	}

	return metadata
}

func parseMetadata(obj *object.FileObject, metadata map[string]interface{}, res *response.Response) {
	for k, v := range metadata {
		k = strings.ToLower(k)
		switch k {
		case "cache-control":
			res.Set(k, v.(string))

		}

		if strings.HasPrefix(k, "x-") {
			res.Set(k, v.(string))
		}
	}

	switch obj.Storage.Kind {
	case "s3":
		for k, v := range metadata {
			switch k {
			case "cache-control", "content-type", "content-encoding", "content-language", "content-disposition":
				res.Set(k, v.(string))
			default:
				res.Set(strings.Join([]string{"x-amz-meta", k}, "-"), v.(string))

			}

		}
	}

}

func inc(obj *object.FileObject, method string) {
	monitoring.Report().Inc(fmt.Sprintf("storage_request;method:%s,storage:%s,bucket:%s,object_type:%s",
		method, obj.Storage.Kind, obj.Storage.Bucket, obj.Type()))
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
