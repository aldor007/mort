package config

import "regexp"

type PresetsYaml struct {
	Quality int `yaml:"quality"`
	Format     string `yaml:"format"`
	Filters struct {
		Thumbnail struct {
			Size []int  `yaml:"size"`
			Mode string `yaml:"mode"`
		} `yaml:"thumbnail"`
		Interlace bool  `yaml:"interlace"`
		Crop struct {
			Size  []int  `yaml:"size"`
			Start []int  `yaml:"start"`
			Mode  string `yaml:"mode"`
		} `yaml:"crop"`
		SmartCrop struct {
			Size []int  `yaml:"size"`
			Mode string `yaml:"mode"`
		} `yaml:"entropy_crop"`
		AutoRotate  bool `yaml:"auto_rtate"`
		Strip      bool `yaml:"strip"`
		Blur   struct {
			Sigma   float64 `yaml:"sigma"`
			MinAmpl float64 `yaml:"minAmpl"`
		} `yaml:"blur"`
		Watermark struct {
			Image    string `yaml:"image"`
			Position string `yaml:"position"`
			Opacity  float32 `yaml:"opacity"`

		}
	} `yaml:"filters"`
}

type TransformYaml struct {
	Path          string `yaml:"path"`
	ParentStorage string `yaml:"parentStorage"`
	ParentBucket  string `yaml:"parentBucket"`
	PathRegexp    *regexp.Regexp
	Kind          string                 `yaml:"kind"`
	Presets       map[string]PresetsYaml `yaml:"presets"`
	CheckParent bool `yaml:"checkParent"`
	ResultKey  string `yaml:"resultKey"`
}

type Storage struct {
	RootPath        string            `yaml:"rootPath,omitempty"`
	Kind            string            `yaml:"kind"`
	Url             string            `yaml:"url,omitempty"`
	Headers         map[string]string `yaml:"headers,omitempty"`
	AccessKey       string            `yaml:"accessKey,omitempty"`
	SecretAccessKey string            `yaml:"secretAccessKey,omitempty"`
	Region          string            `yaml:"region,omitempty"`
	Endpoint        string            `yaml:"endpoint,omitempty"`
	PathPrefix      string            `yaml:"pathPrefix,omitempty"`
	AllowMetadata   bool              `yaml:"allowMetadata,omitempty"`
	Hash            string
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
	Transform *TransformYaml `yaml:"transform,omitempty"`
	Storages  StorageTypes   `yaml:"storages"`
	Keys      []S3Key        `yaml:"keys"`
	Name      string
}

type HeaderYaml struct {
	StatusCodes []int             `yaml:"statusCodes"`
	Values      map[string]string `yaml:"values"`
}
