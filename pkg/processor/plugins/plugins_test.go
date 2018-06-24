package plugins

import (
	"github.com/aldor007/mort/pkg/response"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"net/http"
	"testing"
)

func TestNewPluginsManager(t *testing.T) {
	configStr := `
    compress:
       gzip:
          level: 5
`

	var config map[string]interface{}
	err := yaml.Unmarshal([]byte(configStr), &config)
	if err != nil {
		panic(err)
	}

	pm := NewPluginsManager(config)

	assert.Equal(t, len(pm.list), 1)
}

func TestNewPluginsManagerPanic(t *testing.T) {
	configStr := `
    compress1:
       gzip:
          level: 5
`

	var config map[string]interface{}
	err := yaml.Unmarshal([]byte(configStr), &config)
	if err != nil {
		panic(err)
	}

	assert.Panics(t, func() {
		NewPluginsManager(config)
	})
}

func TestPluginsManager_PreProcess(t *testing.T) {
	configStr := `
    compress:
       gzip:
          level: 5
`

	var config map[string]interface{}
	err := yaml.Unmarshal([]byte(configStr), &config)
	if err != nil {
		panic(err)
	}

	pm := NewPluginsManager(config)
	req, _ := http.NewRequest("GET", "http://mort/local/small.jpg-m", nil)
	req.Header.Add("Accept-Encoding", "gzip")
	body := make([]byte, 1200)
	body[33] = 'a'
	body[324] = 'c'
	res := response.NewBuf(200, body)
	res.Headers.Add("Content-Type", "text/html")

	pm.PreProcess(nil, req)
	pm.PostProcess(nil, req, res)

	assert.Equal(t, len(res.Headers), 3)
	assert.Equal(t, res.Headers.Get("Content-Encoding"), "gzip")
	assert.Equal(t, res.Headers.Get("Vary"), "Accept-Encoding")
}
