package config

import (
	"sync"
	"io/ioutil"
	"gopkg.in/yaml.v2"
)

type Config struct {
	LiipConfig map[string] LiipFiltersYAML
	LocalFilesPath string `yaml:"localFileePath"`
}

type internalConfig  struct {
	LiipConfigPath string `yaml:"liipConfigPath"`
	LiipConfig LiipConfigYAML
	LocalFilesPath string `yaml:"localFilesPath"`
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

	internal := internalConfig{}
	errYaml := yaml.Unmarshal([]byte(data), &internal)
	if errYaml != nil {
		panic(errYaml)
	}


	data, err = ioutil.ReadFile(internal.LiipConfigPath)
	if err != nil {
		panic(err)
	}

	errYaml = yaml.Unmarshal([]byte(data), &internal.LiipConfig)
	if errYaml != nil {
		panic(errYaml)
	}

	self.LiipConfig = internal.LiipConfig.LiipImagine.FilterSets
	self.LocalFilesPath = internal.LocalFilesPath
}