package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyString(t *testing.T) {
	c := Config{}
	err := c.LoadFromString("")
	assert.Nil(t, err)
}

func TestInvalidParentBucketInTransform(t *testing.T) {
	c := Config{}
	err := c.Load("testdata/invalid-parent-bucket.yml")
	assert.NotNil(t, err)
}

func TestInvalidParentStorageInTransform(t *testing.T) {
	c := Config{}
	err := c.Load("testdata/invalid-parent-storage.yml")
	assert.NotNil(t, err)
}

func TestNoBasicStorage(t *testing.T) {
	c := Config{}
	err := c.Load("testdata/no-basic-storage.yml")
	assert.NotNil(t, err)
}
