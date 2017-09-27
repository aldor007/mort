package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"sync"
)

type Config struct {
	Buckets map[string] Bucket `yaml:"buckets"`
}

var instance *Config
var once sync.Once

func GetInstance() *Config {
	once.Do(func() {
		instance = &Config{}
	})
	return instance
}

func (self *Config) Init(filePath string) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	errYaml := yaml.Unmarshal([]byte(data), self)

	if errYaml != nil {
		panic(errYaml)
	}

}
