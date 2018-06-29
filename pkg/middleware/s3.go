package middleware

import (
	"encoding/xml"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/monitoring"
	"github.com/aldor007/mort/pkg/response"

	"github.com/aldor007/go-aws-auth"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

// authHeaderRegexpv4 regular expression for AWS Auth v4 header mode
var autHeaderRegexpv4 = regexp.MustCompile("^(:?[A-Za-z0-9-]+) Credential=(:?.+),\\s*SignedHeaders=(:?[a-zA-Z0-9;-]+),\\s*Signature=(:?[a-zA-Z0-9]+)$")

// authHeaderRegexpv2 regular expression for Aws Auth v2 header mode
var authHeaderRegexpv2 = regexp.MustCompile("^AWS ([A-Za-z0-9-]+):(.+)$")

type s3Context string

// S3AuthCtxKey flag if we have perform authorisation
var S3AuthCtxKey s3Context = "s3-auth"

func isAuthRequired(req *http.Request, auth string, path string) bool {
	method := req.Method
	switch method {
	case "GET", "HEAD", "OPTIONS":
		if path == "/" {
			return true
		}

		if auth != "" {
			return true
		}

		if req.URL.Query().Get("X-Amz-Signature") != "" {
			return true
		}

		return false
	case "POST", "PUT", "DELETE", "PATCH":
		return true
	}

	return true
}

// S3Auth middleware for performing validation of signed requests
type S3Auth struct {
	mortConfig *config.Config // config for buckets
}

// NewS3AuthMiddleware returns S3 compatible authorization handler
// it can handle AWS v2 (S3 mode) and AWS v4 (only header mode without streaming)
func NewS3AuthMiddleware(mortConfig *config.Config) *S3Auth {
	return &S3Auth{mortConfig: mortConfig}
}

// Handler main method of S3AuthMiddleware it check if request should be signed. If so it create copy of request
// and calculate signature and compare result with user request if signature is correct request is passed to next handler
// otherwise it return 403
func (s *S3Auth) Handler(next http.Handler) http.Handler {
	fn := func(resWriter http.ResponseWriter, req *http.Request) {
		path := req.URL.Path
		auth := req.Header.Get("Authorization")

		if !isAuthRequired(req, auth, path) {
			next.ServeHTTP(resWriter, req)
			return
		}

		pathSlice := strings.Split(path, "/")
		pathSliceLen := len(pathSlice)
		if pathSliceLen < 2 {
			monitoring.Log().Warn("S3Auth invalid path")
			res := response.NewString(400, "invalid path")
			res.Send(resWriter)
			return
		}

		bucketName := pathSlice[1]

		var accessKey string
		var signedHeaders []string
		var authAlg string

		matches := autHeaderRegexpv4.FindStringSubmatch(auth)
		if len(matches) == 5 {
			authAlg = "v4"
			reqCredField := matches[2]
			accessKey = strings.Split(reqCredField, "/")[0]
			signedHeaders = strings.Split(matches[3], ";")
		}

		matches = authHeaderRegexpv2.FindStringSubmatch(auth)
		if len(matches) == 3 {
			authAlg = "s3"
			accessKey = matches[1]
		}

		if req.URL.Query().Get("X-Amz-Signature") != "" {
			s.authByQuery(resWriter, req, bucketName, next)
			return
		}

		credential, ok := s.getCredentials(bucketName, accessKey, resWriter)
		if !ok {
			return
		}

		validiatonReq, err := http.NewRequest(req.Method, req.RequestURI, req.Body)
		if err != nil {
			res := response.NewString(401, "")
			monitoring.Log().Error("S3Auth unable to create validation req", zap.Error(err))
			res.Send(resWriter)
			return
		}

		for h, v := range req.Header {
			if strings.HasPrefix(strings.ToLower(h), "x-amz") {
				validiatonReq.Header.Set(h, v[0])
			}

			switch h {
			case "Content-Type", "Content-Md5", "Host", "Date":
				validiatonReq.Header.Set(h, v[0])
			}
		}

		for _, h := range signedHeaders {
			validiatonReq.Header.Set(h, req.Header.Get(h))
		}

		// FIXME: there will be problem with escaped paths
		validiatonReq.URL = req.URL
		validiatonReq.Method = req.Method
		validiatonReq.Body = req.Body
		validiatonReq.Host = req.Host

		if authAlg == "s3" {
			awsauth.SignS3(validiatonReq, credential)
		} else {
			awsauth.Sign4ForRegion(validiatonReq, "mort", "s3", signedHeaders, credential)
		}

		if auth == validiatonReq.Header.Get("Authorization") {
			req.Body = validiatonReq.Body
			if path == "/" {
				s.listAllMyBuckets(resWriter, accessKey)
				return
			}

			ctx := context.WithValue(req.Context(), S3AuthCtxKey, true)

			next.ServeHTTP(resWriter, req.WithContext(ctx))
			return

		}

		monitoring.Log().Warn("S3Auth signature mismatch", zap.String("req.path", req.URL.Path), zap.String("req.method", req.Method))
		response.NewNoContent(403).Send(resWriter)
		return
	}

	return http.HandlerFunc(fn)
}
func (s *S3Auth) getCredentials(bucketName, accessKey string, w http.ResponseWriter) (awsauth.Credentials, bool) {
	var credential awsauth.Credentials
	bucket, ok := s.mortConfig.Buckets[bucketName]
	if !ok {
		buckets := s.mortConfig.BucketsByAccessKey(accessKey)
		if len(buckets) == 0 {
			monitoring.Log().Warn("S3Auth no bucket for access key")
			res := response.NewString(403, "")
			res.Send(w)
			return credential, false
		}

		bucket = buckets[0]
	}

	keys := bucket.Keys
	for _, key := range keys {
		if accessKey == key.AccessKey {
			credential.AccessKeyID = accessKey
			credential.SecretAccessKey = key.SecretAccessKey
			break
		}

	}
	if credential.AccessKeyID == "" {
		res := response.NewString(401, "")
		monitoring.Log().Warn("S3Auth invalid bucket config no access key or invalid", zap.String("bucket", bucketName))
		res.Send(w)
		return credential, false
	}

	return credential, true
}

func (s *S3Auth) listAllMyBuckets(resWriter http.ResponseWriter, accessKey string) {
	type bucketXML struct {
		XMLName      xml.Name `xml:"Bucket"`
		Name         string   `xml:"Name"`
		CreationDate string   `xml:"CreationDate"`
	}

	type listAllBucketsResult struct {
		XMLName xml.Name `xml:"ListAllMyBucketsResult"`
		Owner   struct {
			ID          string `xml:"ID"`
			DisplayName string `xml:"DisplayName"`
		} `xml:"owner"`
		Buckets []bucketXML `xml:"Buckets>Bucket"`
	}

	buckets := s.mortConfig.BucketsByAccessKey(accessKey)
	listAllBucketsXML := listAllBucketsResult{}
	listAllBucketsXML.Owner.DisplayName = accessKey
	listAllBucketsXML.Owner.ID = accessKey

	for _, bucket := range buckets {
		b := bucketXML{}
		b.Name = bucket.Name
		b.CreationDate = time.Now().Format(time.RFC3339)
		listAllBucketsXML.Buckets = append(listAllBucketsXML.Buckets, b)
	}

	b, err := xml.Marshal(listAllBucketsXML)
	if err != nil {
		res := response.NewError(500, err)
		res.Send(resWriter)
		return
	}

	res := response.NewBuf(200, b)
	res.SetContentType("application/xml")
	res.Send(resWriter)
}

func (s *S3Auth) authByQuery(resWriter http.ResponseWriter, r *http.Request, bucketName string, next http.Handler) {
	validationReq := *r
	mortConfig := s.mortConfig

	validationReq.URL.Query().Del("X-Amz-Signature")
	var credential awsauth.Credentials
	accessKey := strings.Split(validationReq.URL.Query().Get("X-Amz-Credential"), "/")[0]

	bucket, ok := mortConfig.Buckets[bucketName]
	if !ok {
		buckets := mortConfig.BucketsByAccessKey(accessKey)
		if len(buckets) == 0 {
			monitoring.Log().Warn("S3Auth no bucket for access key")
			res := response.NewString(403, "")
			res.Send(resWriter)
			return
		}

		bucket = buckets[0]
	}

	if r.URL.Query().Get("X-Amz-Credential") == "" || r.URL.Query().Get("X-Amz-Date") == "" {
		res := response.NewString(401, "")
		monitoring.Log().Warn("S3Auth invalid request no x-amz-credential in query string", zap.String("bucket", bucketName))
		res.Send(resWriter)
		return
	}

	keys := bucket.Keys
	for _, key := range keys {
		if accessKey == key.AccessKey {
			credential.AccessKeyID = accessKey
			credential.SecretAccessKey = key.SecretAccessKey
			break
		}

	}

	if credential.AccessKeyID == "" {
		res := response.NewString(401, "")
		monitoring.Log().Warn("S3Auth invalid bucket config no access key or invalid", zap.String("bucket", bucketName))
		res.Send(resWriter)
		return
	}

	awsauth.PreSign(&validationReq, "mort", "s3", strings.Split(validationReq.URL.Query().Get("X-Amz-SignedHeaders"), ","), credential)

	if validationReq.URL.Query().Get("X-Amz-Signature") == r.URL.Query().Get("X-Amz-Signature") {
		ctx := context.WithValue(r.Context(), S3AuthCtxKey, true)

		next.ServeHTTP(resWriter, r.WithContext(ctx))
		return
	}

	monitoring.Log().Warn("S3Auth signature mismatch", zap.String("req.path", r.URL.Path))
	response.NewNoContent(403).Send(resWriter)
	return

}
