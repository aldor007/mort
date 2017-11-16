package http

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/aldor007/stow"
)

// Kind represents the name of the location/storage type.
const Kind = "http"

const (
	// ConfigAuthType is an optional argument that defines whether to use an IAM role or access key based auth
	ConfigUrl = "url"

	// ConfigAccessKeyID is one key of a pair of AWS credentials.
	ConfigHeader = "headers"
)

func init() {
	makefn := func(config stow.Config) (stow.Location, error) {

		url, ok := config.Config(ConfigUrl)
		if !ok {
			return nil, errors.New("missing url")
		}

		var headers map[string]string

		headersStr, ok := config.Config(ConfigHeader)
		if !ok {
			headers = make(map[string]string)
		} else {
			json.Unmarshal([]byte(headersStr), &headers)
		}

		client := &http.Client{
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).Dial,
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		}

		// Create a location with given config and client (s3 session).
		loc := &location{
			config:   config,
			client:   client,
			endpoint: url,
			headers:  headers,
		}

		return loc, nil
	}

	kindfn := func(u *url.URL) bool {
		return u.Scheme == Kind
	}

	stow.Register(Kind, makefn, kindfn)
}
