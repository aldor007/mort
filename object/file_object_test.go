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

	if obj.HasParent()  {
		t.Errorf("Obj shouldn't have parent")
	}

	if obj.Storage.Kind != "local" {
		t.Errorf("obj should have storage with kind of local")
	}

}

func TestNewFileObjectTransform(t *testing.T) {
	mortConfig := config.GetInstance()
	mortConfig.Load("testdata/bucket-transform.yml")
	obj, err := NewFileObject("/bucket/blog_small/bucket/parent.jpg", mortConfig)

	if err != nil  {
		 t.Errorf("Unexpected to have error when parsing path")
	}

	if obj == nil {
		t.Errorf("Obj shouldn't be nil")
	}

	if !obj.HasParent()  {
		t.Errorf("Obj should have parent")
		t.FailNow()
	}

	parent := obj.Parent
	if parent.Key != "/parent.jpg" {
		t.Errorf("Invalid parent key %s", parent.Key)

	}

	if parent.HasParent() {
		t.Errorf("Parent shouldn't have parent")
	}

	if !obj.HasTransform() {
		t.Errorf("Object should have transformss")
	}

	transCfg := obj.Transforms.BimgOptions()

	if transCfg.Width != 100 {
		t.Errorf("Transform should have 100 px on witdh but has %s", transCfg.Width)
	}

	if transCfg.Height != 100 {
		t.Errorf("Transform should have 100 px on witdh but has %s", transCfg.Width)
	}

}

func TestNewFileObjectTransformOnlyWitdh(t *testing.T) {
	mortConfig := config.GetInstance()
	mortConfig.Load("testdata/bucket-transform.yml")
	obj, err := NewFileObject("/bucket/width/bucket/parent.jpg", mortConfig)

	if err != nil  {
		t.Errorf("Unexpected to have error when parsing path")
	}

	if obj == nil {
		t.Errorf("Obj shouldn't be nil")
	}

	transCfg := obj.Transforms.BimgOptions()

	if transCfg.Width != 100 {
		t.Errorf("Transform should have 100 px on witdh but has %s", transCfg.Width)
	}

	if transCfg.Height != 0 {
		t.Errorf("Transform should have 100 px on witdh but has %s", transCfg.Width)
	}

}

func TestNewFileObjecWithNestedParent(t *testing.T) {
	mortConfig := config.GetInstance()
	mortConfig.Load("testdata/bucket-transform.yml")
	obj, err := NewFileObject("/bucket/width/bucket/height/parent.jpg", mortConfig)

	if err != nil  {
		t.Errorf("Unexpected to have error when parsing path")
	}

	if obj == nil {
		t.Errorf("Obj shouldn't be nil")
	}

	if !obj.HasParent()  {
		t.Errorf("Obj should have parent")
		t.FailNow()
	}

	parent := obj.Parent

	if !parent.HasParent() {
		t.Errorf("Parent shouldn't have parent")
	}

	if parent.Parent.Key != "/parent.jpg" {
		t.Errorf("Parent should have parent /parent.jpg %s", parent.Parent.Key)
	}

}
