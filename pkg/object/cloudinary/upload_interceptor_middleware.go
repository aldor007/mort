package cloudinary

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/monitoring"
	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/response"
	"github.com/aldor007/mort/pkg/storage"
	"go.uber.org/zap"
)

type (
	UploadInterceptorMiddleware struct {
		mortConfig *config.Config // config for buckets
	}

	// UploadResult image success response struct.
	uploadResult struct {
		AssetID  string `json:"asset_id"`
		PublicID string `json:"public_id"`
		Version  int    `json:"version"`
		// VersionID        string          `json:"version_id"`
		// Signature        string          `json:"signature"`
		Format       string `json:"format"`
		ResourceType string `json:"resource_type"`
		// CreatedAt        time.Time       `json:"created_at"`
		Bytes int    `json:"bytes"`
		Type  string `json:"type"`
		// Etag             string          `json:"etag"`
		// URL              string          `json:"url"`
		// SecureURL        string          `json:"secure_url"`
		// AccessMode       string          `json:"access_mode"`
		// Overwritten      bool            `json:"overwritten"`
		OriginalFilename string `json:"original_filename"`
	}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewUploadInterceptorMiddleware(cfg *config.Config) *UploadInterceptorMiddleware {
	return &UploadInterceptorMiddleware{mortConfig: cfg}
}

func verifySignature(params url.Values, apiKey string, secret string) error {
	requestApiKey := params.Get("api_key")
	requestSignature := params.Get("signature")
	if requestApiKey != apiKey {
		return errors.New("api_key does not match")
	}
	requestTimestamp, err := strconv.ParseInt(params.Get("timestamp"), 10, 64)
	if err != nil {
		return errors.New("timestamp is not a number")
	}
	if time.Now().Before(time.Unix(requestTimestamp, 0).Add(-10*time.Minute)) || time.Unix(requestTimestamp, 0).Add(time.Hour).Before(time.Now()) {
		return errors.New("signature expired")
	}
	signedParams := make(url.Values)
	for k := range params {
		if k == "api_key" || k == "signature" {
			continue
		}
		signedParams[k] = params[k]
	}

	encodedUnescapedParams, err := url.QueryUnescape(signedParams.Encode())
	if err != nil {
		return errors.New("failed to encode params for signature verification")
	}
	hash := sha1.New()
	hash.Write([]byte(encodedUnescapedParams + secret))
	if hex.EncodeToString(hash.Sum(nil)) != requestSignature {
		return errors.New("signature does not match")
	}
	return nil
}

func generateRandomID(length int) string {
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}

// Handler intercepts Upload request buckets with 'cloudinary' transformation enabled.
// The upload is performed to Basic storage.
func (u UploadInterceptorMiddleware) Handler(next http.Handler) http.Handler {
	fn := func(resWriter http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			next.ServeHTTP(resWriter, req)
			return
		}

		pathSlice := strings.Split(req.URL.Path, "/")
		if len(pathSlice) < 2 {
			next.ServeHTTP(resWriter, req)
			return
		}
		bucketName := pathSlice[1]
		bucket, ok := u.mortConfig.Buckets[bucketName]
		if !ok || bucket.Transform == nil || bucket.Transform.Kind != Kind {
			next.ServeHTTP(resWriter, req)
			return
		}
		if len(bucket.Keys) != 1 {
			monitoring.Log().Error(
				"CloudinaryUploadInterceptorMiddleware - missing api key and secret",
			)
			res := response.NewString(500, "Server configuration error")
			res.Send(resWriter)
			return
		}
		err := req.ParseMultipartForm(32 << 20) // maxMemory 32MB
		if err != nil {
			res := response.NewString(400, err.Error())
			res.Send(resWriter)
			return
		}
		if err := verifySignature(req.MultipartForm.Value, bucket.Keys[0].AccessKey, bucket.Keys[0].SecretAccessKey); err != nil {
			values, _ := json.Marshal(req.MultipartForm.Value)
			monitoring.Log().Info(
				"CloudinaryUploadInterceptorMiddleware",
				zap.Error(err),
				zap.String("bucket", bucketName),
				zap.String("accessKey", bucket.Keys[0].AccessKey),
				zap.ByteString("values", values),
			)
			res := response.NewString(403, "signature does not match")
			res.Send(resWriter)
			return
		}
		if len(req.MultipartForm.File) != 1 {
			res := response.NewString(400, "only upload of one file is supported")
			res.Send(resWriter)
			return
		}
		fileMeta := req.MultipartForm.File["file"][0]
		fileBody, err := fileMeta.Open()
		if err != nil {
			monitoring.Log().Warn(
				"CloudinaryUploadInterceptorMiddleware - failed to open uploaded file",
				zap.String("bucket", bucketName),
				zap.Error(err),
			)
			res := response.NewString(400, "failed to read file content")
			res.Send(resWriter)
			return
		}
		fileFormat := strings.Split(req.MultipartForm.File["file"][0].Header.Get("Content-Type"), "/")
		if len(fileFormat) != 2 || fileFormat[0] != "image" {
			res := response.NewString(400, "invalid file Content-Type")
			res.Send(resWriter)
			return
		}
		cloudinaryUploadResult := uploadResult{
			AssetID:          generateRandomID(8),
			PublicID:         generateRandomID(20),
			Version:          int(time.Now().Unix()),
			Format:           fileFormat[1],
			ResourceType:     "image",
			Type:             "upload",
			Bytes:            int(fileMeta.Size),
			OriginalFilename: fileMeta.Filename,
		}
		uploadedObj := &object.FileObject{
			Key:     cloudinaryUploadResult.PublicID,
			Uri:     &url.URL{Path: req.URL.Path},
			Storage: bucket.Storages.Basic(),
			Bucket:  bucketName,
			Ctx:     req.Context(),
		}
		resp := storage.Set(
			uploadedObj,
			http.Header{"Content-Type": req.MultipartForm.File["file"][0].Header["Content-Type"]},
			fileMeta.Size,
			fileBody,
		)
		defer fileBody.Close()
		defer resp.Close()
		if resp.StatusCode > 399 {
			res := response.NewString(resp.StatusCode, "parent storage error")
			res.Send(resWriter)
			return
		}
		cloudinaryUploadResultSerializer, err := json.Marshal(cloudinaryUploadResult)
		if err != nil {
			res := response.NewString(500, "failed to prepare response")
			res.Send(resWriter)
			return
		}
		res := response.NewBuf(200, cloudinaryUploadResultSerializer)
		res.Headers.Set("Content-Type", "application/json")
		res.Send(resWriter)
	}
	return http.HandlerFunc(fn)
}
