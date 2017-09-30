package object

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"mort/config"
)

func TestNewFileObjectWhenUnknowBucket(t *testing.T) {
	mortConfig := config.GetInstance()
	_, err := NewFileObject("/bucket/path", mortConfig)

	assert.NotNil(t, err)
}

func TestNewFileObjectNoTransform(t *testing.T) {
	mortConfig := config.GetInstance()
	mortConfig.Load("testdata/bucket-no-transform.yml")
	obj, err := NewFileObject("/bucket/path", mortConfig)

	assert.Nil(t, err)

	assert.NotNil(t, obj)

	assert.False(t, obj.HasParent(), "obj shouldn't have parent")

	assert.Equal(t, "local", obj.Storage.Kind, "obj should have storage with kind of local")

}

func TestNewFileObjectTransform(t *testing.T) {
	mortConfig := config.GetInstance()
	mortConfig.Load("testdata/bucket-transform.yml")
	obj, err := NewFileObject("/bucket/blog_small/bucket/parent.jpg", mortConfig)

	assert.Nil(t, err, "Unexpected to have error when parsing path")

	assert.NotNil(t, obj, "obj should be nil")

	assert.True(t, obj.HasParent(), "obj should have parent")

	parent := obj.Parent

	assert.Equal(t, "/parent.jpg", parent.Key, "invalid parent key")

	assert.False(t, parent.HasParent(), "parent should't have parent")

	assert.True(t, obj.HasTransform(), "obj should have transform")

	transCfg := obj.Transforms.BimgOptions()

	assert.Equal(t, 100, transCfg.Width, "invalid width for transform")

	assert.Equal(t, 100, transCfg.Height, "invalid height for transform")

}

func TestNewFileObjectTransformOnlyWitdh(t *testing.T) {
	mortConfig := config.GetInstance()
	mortConfig.Load("testdata/bucket-transform.yml")
	obj, err := NewFileObject("/bucket/width/bucket/parent.jpg", mortConfig)

	assert.Nil(t, err, "Unexpected to have error when parsing path")

	assert.NotNil(t, obj, "obj should be nil")

	assert.True(t, obj.HasParent(), "obj should have parent")

	parent := obj.Parent

	assert.Equal(t, "/parent.jpg", parent.Key, "invalid parent key")

	assert.False(t, parent.HasParent(), "parent should't have parent")

	assert.True(t, obj.HasTransform(), "obj should have transform")

	transCfg := obj.Transforms.BimgOptions()

	assert.Equal(t, 100, transCfg.Width, "invalid width for transform")

	assert.Equal(t, 0, transCfg.Height, "invalid height for transform")
}

func TestNewFileObjecWithNestedParent(t *testing.T) {
	mortConfig := config.GetInstance()
	mortConfig.Load("testdata/bucket-transform.yml")
	obj, err := NewFileObject("/bucket/width/bucket/height/bucket/parent.jpg", mortConfig)

	assert.Nil(t, err, "Unexpected to have error when parsing path")

	assert.NotNil(t, obj, "obj should be nil")

	assert.True(t, obj.HasParent(), "obj should have parent")

	parent := obj.Parent

	assert.True(t, parent.HasParent(), "parent should have parent")

	assert.Equal(t, "/parent.jpg", parent.Parent.Key, "parent of parent should have correct path")
}
