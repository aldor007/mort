package s3
//
//import (
//	"net/http"
//	"net/url"
//	"time"
//
//
//// Kind represents the name of the location/storage type.
//const Kind = "http"
//
//const (
//	// ConfigAuthType is an optional argument that defines whether to use an IAM role or access key based auth
//	ConfigUrl = "url"
//
//	// ConfigAccessKeyID is one key of a pair of AWS credentials.
//	ConfigHeader = "headers"
//
//)
//
//func init() {
//
//	makefn := func(config stow.Config) (stow.Location, error) {
//
//		url, ok := config.Config(ConfigUrl)
//		if !ok {
//	k		return nil, errors.New("missing url")
//		}
//
//		// Create a new client (s3 session)
//		client, endpoint, err := newS3Client(config)
//		if err != nil {
//			return nil, err
//		}
//
//		// Create a location with given config and client (s3 session).
//		loc := &location{
//			config:         config,
//			client:         client,
//			customEndpoint: endpoint,
//		}
//
//		return loc, nil
//	}
//
//	kindfn := func(u *url.URL) bool {
//		return u.Scheme == Kind
//	}
//
//	stow.Register(Kind, makefn, kindfn)
//}
//
