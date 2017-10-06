package mort

import (
	"net/http"
	"mort/config"
	"github.com/labstack/echo"
	"github.com/crunchytom/go-aws-auth"
	"regexp"
	"strings"
	//"fmt"
)

var AutHeaderRegexpv4 = regexp.MustCompile("^(:?[A-Za-z0-9-]+) Credential=(:?.+),\\s*SignedHeaders=(:?[a-zA-Z0-9;-]+),\\s*Signature=(:?[a-zA-Z0-9]+)$")
var AuthHEaderRegexpv2 = regexp.MustCompile("^AWS ([A-Za-z0-9-]+):(.+)$")

// BasicAuthWithConfig returns an BasicAuth middleware with config.
// See `BasicAuth()`.
func S3Middleware(config *config.Config) echo.MiddlewareFunc {

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			path := req.URL.Path
			pathSlice := strings.Split(path, "/")
			//fmt.Printf("slice = %s path = %s len = %d path= %s \n", pathSlice, path, len(pathSlice), pathSlice[0])
			if len(pathSlice) < 3 {
				return echo.NewHTTPError(400, "invalid path")
			}
			bucketName := pathSlice[1]
			realPath := strings.Join(pathSlice[2:], "/")
			bucket, ok := config.Buckets[bucketName]
			if !ok {
				return echo.NewHTTPError(400, "unknown bucket")
			}

			// TODO: auth for get request
			if req.Method == "GET" && realPath != "/" {
				return next(c)
			}


			auth := req.Header.Get(echo.HeaderAuthorization)
			matches := AutHeaderRegexpv4.FindStringSubmatch(auth)
			if len(matches) == 4 {
				accessKey := matches[1]
				singedHeaders := strings.Split(matches[3], ";")
				//signature := matches[4]
				var credential awsauth.Credentials

				keys := bucket.Keys
				for _, key := range keys {
					if accessKey == key.AccessKey {
						credential.AccessKeyID = accessKey
						credential.SecretAccessKey = key.SecretAccessKey
						break
					}

				}
				if credential.AccessKeyID == "" {
					return echo.ErrForbidden
				}

				valdiatonReq := new(http.Request)
				for h, v := range req.Header {
					if strings.HasPrefix(strings.ToLower(h),"x-amz")  {
						valdiatonReq.Header.Set(h, v[0])
					}
				}

				for _, h := range singedHeaders {
					valdiatonReq.Header.Set(h, req.Header.Get(h))
				}

				valdiatonReq.URL = req.URL
				valdiatonReq.Method = req.Method
				valdiatonReq.Body = req.Body
				valdiatonReq.Host = req.Host

				awsauth.Sign4ForRegion(valdiatonReq, "mort", "s3", credential)
				if auth == valdiatonReq.Header.Get(echo.HeaderAuthorization) {
					return next(c)
				}

				return echo.ErrForbidden
			}

			return echo.ErrForbidden

		}
	}
}