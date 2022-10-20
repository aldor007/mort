package helpers

import (
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

var client = &http.Client{
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
}

// FetchObject download data from given URI
func FetchObject(uri string) ([]byte, error) {
	if strings.HasPrefix(uri, "http") {
		req, err := http.NewRequest("GET", uri, nil)
		if err != nil {
			return nil, err
		}

		response, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		defer response.Body.Close()
		buf, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}

		return buf, nil
	}

	f, err := os.Open(uri)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// IsRangeOrCondition check if request is range or condition
func IsRangeOrCondition(req *http.Request) bool {
	if req.Header.Get("Range") != "" || req.Header.Get("If-Range") != "" {
		return true
	}

	if req.Header.Get("If-Match") != "" || req.Header.Get("If-None-Match") != "" {
		return true
	}

	if req.Header.Get("If-Unmodified-Since") != "" || req.Header.Get("If-Modified-Since") != "" {
		return true
	}

	return false
}
