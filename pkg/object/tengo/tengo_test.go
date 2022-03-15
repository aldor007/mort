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
	assert.Equal(t, fObj.Transforms.HashStr(), "8bb55054d70af2be")

}

func TestParseUsingTengoPreset(t *testing.T) {
	mortConfig := config.Config{}
	mortConfig.BaseConfigPath = "./testdata"
	mortConfig.Load("./testdata/config.yml")

	fObj := object.FileObject{}
	objUrl, _ := url.Parse("https://mort.com/preset-tengo/watermark/image.png")
	fObj.Uri = objUrl
	fObj.Key = objUrl.Path

	parent, err := decodeUsingTengo(objUrl, mortConfig.Buckets["preset-tengo"], &fObj)

	assert.Nil(t, err)
	assert.Equal(t, "/preset-tengo/image.png", parent)
	assert.Equal(t, fObj.Transforms.HashStr(), "50293c54e6375ab9")

}

func TestParseUsingTengoUnknowPreset(t *testing.T) {
	mortConfig := config.Config{}
	mortConfig.BaseConfigPath = "./testdata"
	mortConfig.Load("./testdata/config.yml")

	fObj := object.FileObject{}
	objUrl, _ := url.Parse("https://mort.com/preset-tengo/noname/image.png")
	fObj.Uri = objUrl
	fObj.Key = objUrl.Path

	parent, err := decodeUsingTengo(objUrl, mortConfig.Buckets["preset-tengo"], &fObj)

	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "error: \"unknown preset noname\"")
	assert.Equal(t, parent, "")

}
