package config

import "regexp"

type PresetsYaml struct {
	Quality int `yaml:"quality"`
	Filters struct {
		Thumbnail struct {
			Size []int  `yaml:"size"`
			Mode string `yaml:"mode"`
		} `yaml:"thumbnail"`
		Interlace struct {
			Mode string `yaml:"mode"`
		} `yaml:"interlace"`
		Crop struct {
			Size  []int  `yaml:"size"`
			Start []int  `yaml:"start"`
			Mode  string `yaml:"mode"`
		} `yaml:"crop"`
		SmartCrop struct {
			Size []int  `yaml:"size"`
			Mode string `yaml:"mode"`
		} `yaml:"entropy_crop"`
		AutoRotate interface{} `yaml:"auto_rtate"`
		Strip      interface{} `yaml:"strip"`
	} `yaml:"filters"`
}

type TransformYaml struct {
	Path          string `yaml:"path"`
	ParentStorage string `yaml:"parentStorage"`
	ParentPrefix  string `yaml:"parentPrefix"`
	PathRegexp    *regexp.Regexp
	Kind          string                 `yaml:"kind"`
	Presets       map[string]PresetsYaml `yaml:"presets"`
	Order         struct {
		PresetName int `yaml:"presetName"`
		Parent     int `yaml:"parent"`
	} `yaml:"order"`
}

type Storage struct {
	RootPath        string            `yaml:"rootPath", omitempty`
	Kind            string            `yaml:"kind"`
	Url             string            `yaml:"url",omitempty`
	Headers         map[string]string `yaml:"headers",omitempty`
	AccessKey       string            `yaml:"accessKey",omitempty`
	SecretAccessKey string            `yaml:"secretAccessKey",omitempty`
	Region          string            `yaml:"region",omitempty`
	Endpoint        string            `yaml:"endpoint",omitempty`
}

type StorageTypes map[string]Storage

func (s *StorageTypes) Basic() Storage {
	return s.Get("basic")
}

func (s *StorageTypes) Transform() Storage {
	return s.Get("transform")
}

func (s *StorageTypes) Get(name string) Storage {
	return (*s)[name]
}

type S3Key struct {
	AccessKey       string `yaml:"accessKey"`
	SecretAccessKey string `yaml:"secretAccessKey"`
}

type Bucket struct {
	Transform *TransformYaml `yaml:"transform",omitempty`
	Storages  StorageTypes   `yaml:"storages"`
	Keys      []S3Key        `yaml:"keys"`
}

type HeaderYaml struct {
	StatusCodes []int             `yaml:"statusCodes""`
	Values      map[string]string `yaml:"values"`
}
