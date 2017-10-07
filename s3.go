package mort

import (
	"net/http"
	"mort/config"
	"github.com/labstack/echo"
	//"github.com/crunchytom/go-aws-auth"
	//awsv4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"regexp"
	"strings"
	"errors"
	"github.com/crunchytom/go-aws-auth"
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

			bucket, ok := config.Buckets[bucketName]

			if !ok {
				return echo.NewHTTPError(404, "unknown bucket")
			}

			// TODO: auth for get request


			auth := req.Header.Get(echo.HeaderAuthorization)


			if req.Method == "GET" && auth == "" {
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
					return next(c)
				}

				return echo.ErrForbidden
			}

			return echo.ErrForbidden

		}
	}
}