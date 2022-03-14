package tengo

import (
	"net/url"
	"testing"

	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/object"
	"github.com/stretchr/testify/assert"
)

func TestParseUsingTengo(t *testing.T) {
	mortConfig := config.Config{}
	mortConfig.BaseConfigPath = "./testdata"
	mortConfig.Load("./testdata/config.yml")

	fObj := object.FileObject{}
	objUrl, _ := url.Parse("https://mort.com/bucket/image,w100,h100.png")
	fObj.Uri = objUrl
	fObj.Key = objUrl.Path

	parent, err := decodeUsingTengo(objUrl, mortConfig.Buckets["tengo"], &fObj)

	assert.Nil(t, err)
	assert.Equal(t, "/bucket/image.png", parent)
	assert.True(t, fObj.CheckParent)

}
