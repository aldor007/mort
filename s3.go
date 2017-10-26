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
	"fmt"
)

var AutHeaderRegexpv4 = regexp.MustCompile("^(:?[A-Za-z0-9-]+) Credential=(:?.+),\\s*SignedHeaders=(:?[a-zA-Z0-9;-]+),\\s*Signature=(:?[a-zA-Z0-9]+)$")
var AuthHeaderRegexpv2 = regexp.MustCompile("^AWS ([A-Za-z0-9-]+):(.+)$")

func isAuthRequired(method string, auth string, path string) bool {
	switch method {
	case "GET", "HEAD", "OPTIONS":
		if path == "/" {
			return true
		}

		if auth != "" {
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
func S3AuthMiddleware(mortConfig *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			path := req.URL.Path
			pathSlice := strings.Split(path, "/")
			pathSliceLen := len(pathSlice)
			if pathSliceLen < 2 {
				return echo.NewHTTPError(400, "invalid path")
			}

			bucketName := pathSlice[1]

			auth := req.Header.Get(echo.HeaderAuthorization)
			if !isAuthRequired(req.Method, auth, path) {
				return next(c)
			}

			var accessKey string
			var signedHeaders []string
			var bucket config.Bucket
			var credential awsauth.Credentials
			var authAlg string

			matches := AutHeaderRegexpv4.FindStringSubmatch(auth)
			if len(matches) == 5 {
				authAlg = "v4"
				alg := matches[1]
				if alg != "AWS4-HMAC-SHA256" {
					return echo.NewHTTPError(400, errors.New("invalid algorithm"))
				}

				reqCredField := matches[2]
				accessKey = strings.Split(reqCredField, "/")[0]
				signedHeaders = strings.Split(matches[3], ";")
			}

			matches = AuthHeaderRegexpv2.FindStringSubmatch(auth)
			if len(matches) == 3 {
				authAlg = "s3"
				accessKey = matches[1]
			}

			bucket, ok := mortConfig.Buckets[bucketName]
			if !ok {
				buckets := mortConfig.BucketsByAccessKey(accessKey)
				if len(buckets) == 0 {
					return echo.ErrForbidden
				}

				bucket = buckets[0]
			}
			fmt.Println("aaa ", accessKey)
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

			validiatonReq, err := http.NewRequest(req.Method, req.RequestURI, req.Body)
			if err != nil {
				return echo.ErrUnauthorized
			}

			for h, v := range req.Header {
				if strings.HasPrefix(strings.ToLower(h),"x-amz")  {
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
				awsauth.SignS3(validiatonReq,  credential)
			} else {
				awsauth.Sign4ForRegion(validiatonReq, "mort", "s3", credential)
			}


			if auth == validiatonReq.Header.Get(echo.HeaderAuthorization) {
				req.Body = validiatonReq.Body
				c.Set("accessKey", accessKey)
				if path == "/"  {
					return listAllMyBuckets(c, mortConfig, accessKey)
				}
				return next(c)
			}

			fmt.Printf("auth = %s valid = %s", auth, validiatonReq.Header.Get(echo.HeaderAuthorization))
			return echo.ErrForbidden

		}
	}
}


func listAllMyBuckets(c echo.Context, mortConfig *config.Config, accessKey string) error {
	type bucketXml struct {
		XMLName     xml.Name `xml:"Bucket"`
		Name string `xml:"Name"`
		CreationDate string `xml:"CreationDate"`

	}

	type listAllBucketsResult struct {
		XMLName     xml.Name `xml:"ListAllMyBucketsResult"`
		Owner      struct {
			ID     string `xml:"ID"`
			DisplayName string `xml:"DisplayName"`
		} `xml:"owner"`
		Buckets []bucketXml `xml:"Buckets>Bucket"`
	}

	buckets := mortConfig.BucketsByAccessKey(accessKey)
	listAllBucketsXML := listAllBucketsResult{}
	listAllBucketsXML.Owner.DisplayName = accessKey
	listAllBucketsXML.Owner.ID = accessKey
	//listAllBucketsXML.Buckets = make([]bucketXml, len(buckets))
	for _, bucket := range buckets {
		b := bucketXml{}
		b.Name = bucket.Name
		b.CreationDate = time.Now().Format(time.RFC3339)
		listAllBucketsXML.Buckets = append(listAllBucketsXML.Buckets, b)
	}

	c.XML(200, listAllBucketsXML)
	return nil
}
