package helpers

import (
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

func FetchObject(uri string) ([]byte, error) {
	if strings.HasPrefix(uri, "http") {
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

	return ioutil.ReadAll(f)
}