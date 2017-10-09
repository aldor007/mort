package mort

import (
	"net/http"
	"mort/config"
	"github.com/labstack/echo"
	//"github.com/crunchytom/go-aws-auth"
	//awsv4 "githubBucketcom/aws/aws-sdk-go/aws/signer/v4"
	"regexp"
	"strings"
	"errors"
	"github.com/aldor007/go-aws-auth"
	"encoding/xml"
	"time"
)

var AutHeaderRegexpv4 = regexp.MustCompile("^(:?[A-Za-z0-9-]+) Credential=(:?.+),\\s*SignedHeaders=(:?[a-zA-Z0-9;-]+),\\s*Signature=(:?[a-zA-Z0-9]+)$")
var AuthHeaderRegexpv2 = regexp.MustCompile("^AWS ([A-Za-z0-9-]+):(.+)$")

func isAuthRequired(method string, auth string, path string) bool {
	switch method {
	case "GET", "HEAD", "OPTIONS":
		if auth != "" {
			return true
		}
		if path == "" {
			return true
		}
		return false
	case "POST", "PUT", "DELETE", "PATCH":
		return true
	}

	return true
}

// BasicAuthWithConfig returns an BasicAuth middleware with config.
// See `BasicAuth()`.
func S3AuthMiddleware(config *config.Config) echo.MiddlewareFunc {

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			path := req.URL.Path
			pathSlice := strings.Split(path, "/")
			pathSliceLen := len(pathSlice)
			//fmt.Printf("slice = %s path = %s len = %d path= %s \n", pathSlice, path, len(pathSlice), pathSlice[0])
			if pathSliceLen < 2 {
				return echo.NewHTTPError(400, "invalid path")
			}
			bucketName := pathSlice[1]
			//realPath := ""
			//if pathSliceLen > 2 {
			//	realPath = strings.Join(pathSlice[2:], "/")
			//}


			// TODO: auth for get request

			auth := req.Header.Get(echo.HeaderAuthorization)
			if !isAuthRequired(req.Method, auth, path) {
				return next(c)
			}

			matches := AutHeaderRegexpv4.FindStringSubmatch(auth)
			if len(matches) == 5 {
				alg := matches[1]
				if alg != "AWS4-HMAC-SHA256" {
					return echo.NewHTTPError(400, errors.New("invalid algorithm"))
				}

				reqCredField := matches[2]
				accessKey := strings.Split(reqCredField, "/")[0]
				singedHeaders := strings.Split(matches[3], ";")
				//signature := matches[4]
				var credential awsauth.Credentials
				bucket, ok := config.Buckets[bucketName]

				if !ok {
					buckets := config.BucketsByAccessKey(accessKey)
					if len(buckets) == 0{
						return echo.NewHTTPError(404, "unknown bucket")
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
					return echo.ErrUnauthorized
				}

				validiatonReq, err := http.NewRequest(req.Method, req.URL.String(), req.Body)
				if err != nil {
					return echo.ErrUnauthorized
				}
				for h, v := range req.Header {
					if strings.HasPrefix(strings.ToLower(h),"x-amz")  {
						validiatonReq.Header.Set(h, v[0])
					}

					switch h {
						case "Content-Type", "Content-Md5", "Host":
							validiatonReq.Header.Set(h, v[0])
					}
				}

				for _, h := range singedHeaders {
					validiatonReq.Header.Set(h, req.Header.Get(h))
				}

				validiatonReq.URL = req.URL
				validiatonReq.Method = req.Method
				validiatonReq.Body = req.Body
				validiatonReq.Host = req.Host

				awsauth.Sign4ForRegion(validiatonReq, "mort", "s3", credential)
				if auth == validiatonReq.Header.Get(echo.HeaderAuthorization) {
					c.Set("accessKey", accessKey)
					return next(c)
				}

				return echo.ErrForbidden
			}

			matches = AuthHeaderRegexpv2.FindStringSubmatch(auth)
			if len(matches) == 3 {

				accessKey := matches[1]
				c.Set("accessKey", accessKey)
				return next(c)
				var credential awsauth.Credentials
				bucket, ok := config.Buckets[bucketName]

				if !ok {
					buckets := config.BucketsByAccessKey(accessKey)
					if len(buckets) == 0{
						return echo.NewHTTPError(404, "unknown bucket")
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
					return echo.ErrUnauthorized
				}

				validiatonReq, err := http.NewRequest(req.Method, req.URL.String(), req.Body)
				if err != nil {
					return echo.ErrUnauthorized
				}
				for h, v := range req.Header {
					if strings.HasPrefix(strings.ToLower(h),"x-amz")  {
						validiatonReq.Header.Set(h, v[0])
					}

					switch h {
					case "Content-Type", "Content-Md5", "Host":
						validiatonReq.Header.Set(h, v[0])
					}
				}

				validiatonReq.URL = req.URL
				validiatonReq.Method = req.Method
				validiatonReq.Body = req.Body
				validiatonReq.Host = req.Host
				awsauth.SignS3(validiatonReq,  credential)
				if auth == validiatonReq.Header.Get(echo.HeaderAuthorization) {
					c.Set("accessKey", accessKey)
					return next(c)
				}

				return echo.ErrForbidden
			}

			return echo.ErrForbidden

		}
	}
}

type bucketXml struct {
	XMLName     xml.Name `xml:"Bucket"`
	Name string `xml:"Name"`
	CreationDate string `xml:"CreationDate"`

}

type listAllBucketsResult struct {
	XMLName     xml.Name `xml:"ListBucketsResult"`
	Owner      struct {
		ID     string `xml:"ID"`
		DisplayName string `xml:"DisplayName"`
	} `xml:"owner"`
	Buckets []bucketXml `xml:"Buckets>Bucket"`
}


func S3Middleware(config *config.Config) echo.MiddlewareFunc {

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			accessKey := c.Get("accessKey")
			if accessKey == nil {
				return next(c)
			}

			buckets := config.BucketsByAccessKey(accessKey.(string))
			listAllBucketsXML := listAllBucketsResult{}
			listAllBucketsXML.Owner.DisplayName = "test"
			listAllBucketsXML.Owner.ID = "test"
			//listAllBucketsXML.Buckets = make([]bucketXml, len(buckets))
			for _, bucket := range buckets {
				if bucket.Name != "" {
					b := bucketXml{}
					b.Name = bucket.Name
					b.CreationDate = time.Now().Format(time.RFC3339)
					listAllBucketsXML.Buckets = append(listAllBucketsXML.Buckets, b)
				}
			}

			c.XML(200, listAllBucketsXML)
			return nil
		}
	}
}
