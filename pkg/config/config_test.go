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

func TestInvalidYaml(t *testing.T) {
	c := GetInstance()
	assert.Panics(t, func() {
		c.load([]byte(`
	server:
		a: [
`))
	})
}

func TestInvalidFile(t *testing.T) {
	c := GetInstance()
	assert.Panics(t, func() {
		c.Load("no-file")
	})
}

func TestConfig_Load(t *testing.T) {
	c := Config{}
	c.BaseConfigPath = "testdata"
	err := c.Load("testdata/config.yml")

	assert.Nil(t, err)

	buckets := c.BucketsByAccessKey("acc")

	assert.Equal(t, len(buckets), 1)

	bucket := c.Buckets["media"]
	assert.Equal(t, bucket.Storages.Transform().Kind, "local-meta")
}

func TestConfig_Load_TengoInvalid(t *testing.T) {
	c := Config{}
	err := c.Load("testdata/config-tengo-invalid.yml")

	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "unable to read tengo script file \"\", error open configuration: no such file or directory")
}

func TestConfig_Load_TengoCompileError(t *testing.T) {
	c := Config{}
	c.BaseConfigPath = "testdata"
	err := c.Load("testdata/config-tengo-compile-error.yml")

	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "unable to compile tengo script tengo.tengo error Compile Error: unresolved reference 'aaaa'\n\tat (main):1:1")
}

func TestConfig_Transform_ForParser(t *testing.T) {
	c := Config{}
	c.BaseConfigPath = "testdata"
	err := c.Load("testdata/config.yml")

	assert.Nil(t, err)
	ten := c.Buckets["tengo"].Transform.ForParser()
	assert.Nil(t, ten.TengoScript)
}
