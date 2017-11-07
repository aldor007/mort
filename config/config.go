package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"regexp"
	"sync"
	"fmt"
	"github.com/pkg/errors"

	"mort/log"
)

type Config struct {
	Buckets map[string]Bucket `yaml:"buckets"`
	Headers []HeaderYaml      `yaml:"headers"`
	accessKeyBucket map[string][] string
}

var instance *Config
var once sync.Once
var storageKinds []string = []string{"local", "local-meta", "s3", "http"}

func GetInstance() *Config {
	once.Do(func() {
		instance = &Config{}
	})
	return instance
}

func (self *Config) Load(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	return self.load(data)
}

func (self *Config) LoadFromString(data string) error {
	return self.load([]byte(data))
}

func (self *Config) load(data []byte) error {
	errYaml := yaml.Unmarshal(data, self)
	if errYaml != nil {
		panic(errYaml)
	}

	self.accessKeyBucket = make(map[string][]string)
	for name, bucket := range self.Buckets {
		if bucket.Transform != nil {
			if bucket.Transform.Path != "" {
				bucket.Transform.PathRegexp = regexp.MustCompile(bucket.Transform.Path)
			}

			if bucket.Transform.Order.PresetName == bucket.Transform.Order.Parent {
				bucket.Transform.Order.Parent = 0
				bucket.Transform.Order.PresetName = 1
			}

			if bucket.Transform.ParentStorage == "" {
				bucket.Transform.ParentStorage = "basic"
			}
		}

		bucket.Name = name
		self.Buckets[name] = bucket
		for _, key := range bucket.Keys {
			self.accessKeyBucket[key.AccessKey] = append(self.accessKeyBucket[key.AccessKey], name)
		}
	}

	return self.validate()
}

func (c *Config) BucketsByAccessKey(accessKey string) []Bucket {
	list := c.accessKeyBucket[accessKey]
	var buckets []Bucket = make([]Bucket, len(list))
	for i, name := range list {
		buckets[i] = c.Buckets[name]
	}

	return buckets
}

func configInvalidError(msg string) error {
	log.Log().Warnw(msg)
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
	if transform.Kind != "presets" {
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
