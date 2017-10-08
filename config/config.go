package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"regexp"
	"sync"
)

type Config struct {
	Buckets map[string]Bucket `yaml:"buckets"`
	Headers []HeaderYaml      `yaml:"headers"`
	accessKeyBucket map[string][] string
}

var instance *Config
var once sync.Once

func GetInstance() *Config {
	once.Do(func() {
		instance = &Config{}
	})
	return instance
}

//func (c *Config) validate() {
//	for name, bucket := range c.Buckets {
//		//if bucket.Storages.Basic() == nil {
//		//	panic("No basic storage for " + name)
//		//}
//	}
//}

func (self *Config) Load(filePath string) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	errYaml := yaml.Unmarshal([]byte(data), self)

	self.accessKeyBucket = make(map[string][]string)
	for name, bucket := range self.Buckets {
		if bucket.Transform != nil {
			if bucket.Transform.Path != "" {
				bucket.Transform.PathRegexp = regexp.MustCompile(bucket.Transform.Path)
			}

			if bucket.Transform.ParentStorage == "" {
				bucket.Transform.ParentStorage = "basic"
			}
		}

		bucket.Name = name
		self.Buckets[name] = bucket
		for _, key := range bucket.Keys {
			//if self.accessKeyBucket[key.AccessKey] == nil {
			//	self.accessKeyBucket[key.AccessKey] = make([]stri
			//}

			self.accessKeyBucket[key.AccessKey] = append(self.accessKeyBucket[key.AccessKey], name)
		}
	}

	if errYaml != nil {
		panic(errYaml)
	}

}

func (c *Config) BucketsByAccessKey(accessKey string) []Bucket {
	list := c.accessKeyBucket[accessKey]
	var buckets []Bucket = make([]Bucket, len(list))
	for _, name := range list {
		buckets = append(buckets, c.Buckets[name])
	}

	return buckets
}
