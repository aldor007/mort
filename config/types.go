package config

import "regexp"

// PresetsYaml describe properties of transform preset
type PresetsYaml struct {
	Quality int    `yaml:"quality"`
	Format  string `yaml:"format"`
	Filters struct {
		Thumbnail struct {
			Size []int  `yaml:"size"`
			Mode string `yaml:"mode"`
		} `yaml:"thumbnail"`
		Interlace bool `yaml:"interlace"`
		Crop      struct {
			Size  []int  `yaml:"size"`
			Start []int  `yaml:"start"`
			Mode  string `yaml:"mode"`
		} `yaml:"crop"`
		SmartCrop struct {
			Size []int  `yaml:"size"`
			Mode string `yaml:"mode"`
		} `yaml:"entropy_crop"`
		AutoRotate bool `yaml:"auto_rtate"`
		Strip      bool `yaml:"strip"`
		Blur       struct {
			Sigma   float64 `yaml:"sigma"`
			MinAmpl float64 `yaml:"minAmpl"`
		} `yaml:"blur"`
		Watermark struct {
			Image    string  `yaml:"image"`
			Position string  `yaml:"position"`
			Opacity  float32 `yaml:"opacity"`
		}
	} `yaml:"filters"`
}

// TransformYaml describe transform for bucket
type TransformYaml struct {
	Path          string `yaml:"path"`
	ParentStorage string `yaml:"parentStorage"`
	ParentBucket  string `yaml:"parentBucket"`
	PathRegexp    *regexp.Regexp
	Kind          string                 `yaml:"kind"`
	Presets       map[string]PresetsYaml `yaml:"presets"`
	CheckParent   bool                   `yaml:"checkParent"`
	ResultKey     string                 `yaml:"resultKey"`
}

// Storage contains information about kind of used storage
type Storage struct {
	RootPath        string            `yaml:"rootPath,omitempty"`        // root path for local-* storage
	Kind            string            `yaml:"kind"`                      // type of storage from list ("local", "local-meta", "s3", "http", "noop")
	Url             string            `yaml:"url,omitempty"`             // Url for http storage
	Headers         map[string]string `yaml:"headers,omitempty"`         // request headers for http storage
	AccessKey       string            `yaml:"accessKey,omitempty"`       // access key for s3 storage
	SecretAccessKey string            `yaml:"secretAccessKey,omitempty"` // SecretAccessKey for s3 storage
	Region          string            `yaml:"region,omitempty"`          // region for s3 storage
	Endpoint        string            `yaml:"endpoint,omitempty"`        // endpoint for s3 storage
	PathPrefix      string            `yaml:"pathPrefix,omitempty"`      // prefix in path for all storage
	Hash            string            // unique hash for given storage
}

// StorageTypes contains map of storage for bucket
type StorageTypes map[string]Storage

// Basic return storage that contains originals object
func (s *StorageTypes) Basic() Storage {
	return s.Get("basic")
}

// Transform return strorage in which we should storage processed objects
func (s *StorageTypes) Transform() Storage {
	return s.Get("transform")
}

// Get basic method for getting storage by name
func (s *StorageTypes) Get(name string) Storage {
	return (*s)[name]
}

// S3Key define credentials for s3 auth
type S3Key struct {
	AccessKey       string `yaml:"accessKey"`
	SecretAccessKey string `yaml:"secretAccessKey"`
}

// Bucket describe single bucket entry in config
type Bucket struct {
	Transform *TransformYaml `yaml:"transform,omitempty"`
	Storages  StorageTypes   `yaml:"storages"`
	Keys      []S3Key        `yaml:"keys"`
	Name      string
}

// HeaderYaml allow you to override response headers
type HeaderYaml struct {
	StatusCodes []int             `yaml:"statusCodes"`
	Values      map[string]string `yaml:"values"`
}
