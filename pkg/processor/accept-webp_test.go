package processor

import (
	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/object"
	"github.com/aldor007/mort/pkg/response"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
)

func pathToURL(urlPath string) *url.URL {
	u, _ := url.Parse(urlPath)
	return u
}

func TestWebpInAccept(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://mort/local/small.jpg-m", nil)
	req.Header.Add("Accept", "image/webp")

	mortConfig := config.Config{}
	err := mortConfig.Load("./benchmark/small.yml")
	assert.Nil(t, err)

	obj, err := object.NewFileObject(req.URL, &mortConfig)
	assert.Nil(t, err)

	res := response.NewNoContent(200)

	obj.Ctx = req.Context()
	w := WebpHook{}
	w.preProcess(obj, req)
	w.postProcess(obj, req, res)

	assert.Equal(t, res.Headers.Get("Vary"), "accept")
	assert.Equal(t, obj.Transforms.FormatStr, "webp")
}

func TestDontChangeWhenNoAccept(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://mort/local/small.jpg-m", nil)
	req.Header.Add("Accept", "image/*")

	mortConfig := config.Config{}
	err := mortConfig.Load("./benchmark/small.yml")
	assert.Nil(t, err)

	obj, err := object.NewFileObject(req.URL, &mortConfig)
	assert.Nil(t, err)

	res := response.NewNoContent(200)

	obj.Ctx = req.Context()
	w := WebpHook{}
	w.preProcess(obj, req)
	w.postProcess(obj, req, res)

	assert.Equal(t, res.Headers.Get("Vary"), "")
	assert.Equal(t, obj.Transforms.FormatStr, "")
}
