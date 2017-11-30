package config

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"regexp"
	"strings"
	"sync"

	"github.com/aldor007/mort/log"
)

// Config contains configuration for buckets etc
//
// Config should be used like singleton
type Config struct {
	Buckets         map[string]Bucket `yaml:"buckets"`
	Headers         []HeaderYaml      `yaml:"headers"`
	accessKeyBucket map[string][]string
}

var instance *Config
var once sync.Once
var storageKinds = []string{"local", "local-meta", "s3", "http", "noop"}

// GetInstance return single instance of Config object
func GetInstance() *Config {
	once.Do(func() {
		instance = &Config{}
	})
	return instance
}

// Load reads config data from file
// How configuration file should be formatted see README.md
func (c *Config) Load(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	return c.load(data)
}

// LoadFromString parse configuration form string
func (c *Config) LoadFromString(data string) error {
	return c.load([]byte(data))
}

func (c *Config) load(data []byte) error {
	errYaml := yaml.Unmarshal(data, c)
	if errYaml != nil {
		panic(errYaml)
	}

	c.accessKeyBucket = make(map[string][]string)
	for name, bucket := range c.Buckets {
		if bucket.Transform != nil {
			if bucket.Transform.Path != "" {
				bucket.Transform.PathRegexp = regexp.MustCompile(bucket.Transform.Path)
			}

			if bucket.Transform.ParentStorage == "" {
				bucket.Transform.ParentStorage = "basic"
			}
		}

		for sName, storage := range c.Buckets[name].Storages {
			storage.Hash = name + sName + storage.Kind
			if sName == "transforms" {
				if storage.PathPrefix == "" {
					storage.PathPrefix = "transforms"
				}
			}

			bucket.Storages[sName] = storage
		}

		bucket.Name = name
		c.Buckets[name] = bucket
		for _, key := range bucket.Keys {
			c.accessKeyBucket[key.AccessKey] = append(c.accessKeyBucket[key.AccessKey], name)
		}
	}

	return c.validate()
}

// BucketsByAccessKey return list of buckets that have given accessKey
func (c *Config) BucketsByAccessKey(accessKey string) []Bucket {
	list := c.accessKeyBucket[accessKey]
	buckets := make([]Bucket, len(list))
	for i, name := range list {
		buckets[i] = c.Buckets[name]
	}
	return buckets
}

func configInvalidError(msg string) error {
	log.Logs().Warnw(msg)
	return errors.New(msg)
}

func (c *Config) validateStorage(bucketName string, storages StorageTypes) error {
	validStorageKind := false
	var err error
	basic := storages.Basic()
	if basic.Kind == "" {
		return configInvalidError(fmt.Sprintf("%s basic storage is required", bucketName))
	}

	for storageName, storage := range storages {
		validStorageKind = false
		for _, k := range storageKinds {
			if k == storage.Kind {
				validStorageKind = true
				break
			}
		}
		errorMsgPrefix := fmt.Sprintf("%s has invalid config for storage %s kind %s", bucketName, storageName, storage.Kind)
		if !validStorageKind {
			err = configInvalidError(fmt.Sprintf("%s has invalid storage %s kind %s valid %s", bucketName, storageName,
				storage.Kind, storageKinds))
		}

		if storage.Kind == "local" || storage.Kind == "local-meta" {
			if storage.RootPath == "" {
				err = configInvalidError(fmt.Sprintf("%s - no rootPath", errorMsgPrefix))
			}
		}

		if storage.Kind == "http" {
			if storage.Url == "" {
				err = configInvalidError(fmt.Sprintf("%s - no url", errorMsgPrefix))
			}
		}

		if storage.Kind == "s3" {
			if storage.AccessKey == "" {
				err = configInvalidError(fmt.Sprintf("%s - no accessKey", errorMsgPrefix))
			}

			if storage.SecretAccessKey == "" {
				err = configInvalidError(fmt.Sprintf("%s - no secretAccessKey", errorMsgPrefix))
			}

		}
	}

	return err
}

func (c *Config) validateTransform(bucketName string, bucket Bucket) error {
	transform := bucket.Transform
	var err error
	errorMsgPrefix := fmt.Sprintf("%s has invalid transform config", bucketName)
	if transform.Kind != "presets" && transform.Kind != "query" && transform.Kind != "presets-query" {
		return configInvalidError(fmt.Sprintf("%s - unknown kind %s", errorMsgPrefix, transform.Kind))
	}

	if bucket.Storages.Get(transform.ParentStorage).Kind == "" {
		err = configInvalidError(fmt.Sprintf("%s - no parentStorage of name %s", errorMsgPrefix, transform.ParentStorage))
	}

	if transform.ParentBucket != "" {
		if _, ok := c.Buckets[transform.ParentBucket]; !ok {
			err = configInvalidError(fmt.Sprintf("%s - parentBucket %s doesn't exist", errorMsgPrefix, transform.ParentBucket))
		}
	}

	if transform.Kind == "presets" {
		if strings.Index(transform.Path, "(?P<presetName>") == -1 {
			err = configInvalidError(fmt.Sprintf("%s invalid transform regexp it should have capturing group for presetName `(?P<presetName>``", errorMsgPrefix))
		}

		if strings.Index(transform.Path, "(?P<parent>") == -1 {
			err = configInvalidError(fmt.Sprintf("%s invalid transform regexp it should have capturing group for parent `(?P<parent>``", errorMsgPrefix))
		}
	}

	return err

}

func (c *Config) validate() error {
	for name, bucket := range c.Buckets {
		err := c.validateStorage(name, bucket.Storages)
		if err != nil {
			return err
		}

		if bucket.Transform != nil {
			err = c.validateTransform(name, bucket)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
