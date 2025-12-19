package config

import (
	"regexp"

	"github.com/d5/tengo/v2"
)

// Filter yaml configuration filters
type Filters struct {
	Thumbnail *struct {
		Width               int    `yaml:"width"`
		Height              int    `yaml:"height"`
		Mode                string `yaml:"mode"`
		PreserveAspectRatio bool   `yaml:"preserveAspectRatio"`
		Fill                bool   `yaml:"fill"`
	} `yaml:"thumbnail,omitempty"`
	Interlace bool `yaml:"interlace"`
	Crop      *struct {
		Width   int    `yaml:"width"`
		Height  int    `yaml:"height"`
		Gravity string `yaml:"gravity"`
		Mode    string `yaml:"mode"`
		Embed   bool   `yaml:"embed"`
	} `yaml:"crop,omitempty"`
	Extract *struct {
		Width  int `yaml:"width"`
		Height int `yaml:"height"`
		Top    int `yaml:"top"`
		Left   int `yaml:"left"`
	} `yaml:"extract,omitempty"`
	ResizeCropAuto *struct {
		Width  int `yaml:"width"`
		Height int `yaml:"height"`
	} `yaml:"resizeCropAuto,omitempty"`
	AutoRotate bool `yaml:"auto_rotate"`
	Grayscale  bool `yaml:"grayscale"`
	Strip      bool `yaml:"strip"`
	Blur       *struct {
		Sigma   float64 `yaml:"sigma"`
		MinAmpl float64 `yaml:"minAmpl"`
	} `yaml:"blur,omitempty"`
	Watermark *struct {
		Image    string  `yaml:"image"`
		Position string  `yaml:"position"`
		Opacity  float32 `yaml:"opacity"`
	} `yaml:"watermark,omitempty"`
	Rotate *struct {
		Angle int `yaml:"angle"`
	} `yaml:"rotate,omitempty"`
}

// Preset describe properties of transform preset
type Preset struct {
	Quality int     `yaml:"quality" json:"quality"`
	Format  string  `yaml:"format" json:"format"`
	Filters Filters `yaml:"filters" json:"filters"`
}

// Transform describe transform for bucket
type Transform struct {
	Path          string `yaml:"path"`
	ParentStorage string `yaml:"parentStorage"`
	ParentBucket  string `yaml:"parentBucket"`
	PathRegexp    *regexp.Regexp
	Kind          string            `yaml:"kind"`
	Presets       map[string]Preset `yaml:"presets"`
	CheckParent   bool              `yaml:"checkParent"`
	ResultKey     string            `yaml:"resultKey"`
	TengoPath     string            `yaml:"tengoPath"`
	TengoScript   *tengo.Compiled
}

func (t *Transform) ForParser() *Transform {
	return &Transform{
		Path:          t.Path,
		ParentStorage: t.ParentStorage,
		ParentBucket:  t.ParentBucket,
		PathRegexp:    t.PathRegexp,
		Kind:          t.Kind,
		Presets:       t.Presets,
		CheckParent:   t.CheckParent,
		ResultKey:     t.ResultKey,
	}

}

// Storage contains information about kind of used storage
type Storage struct {
	RootPath           string            `yaml:"rootPath,omitempty"`        // root path for local-* storage
	Kind               string            `yaml:"kind"`                      // type of storage from list ("local", "local-meta", "s3", "http", "b2","noop")
	Url                string            `yaml:"url,omitempty"`             // Url for http storage
	Headers            map[string]string `yaml:"headers,omitempty"`         // request headers for http storage
	AccessKey          string            `yaml:"accessKey,omitempty"`       // access key for s3 storage
	SecretAccessKey    string            `yaml:"secretAccessKey,omitempty"` // SecretAccessKey for s3 storage
	Region             string            `yaml:"region,omitempty"`          // region for s3 storage
	Endpoint           string            `yaml:"endpoint,omitempty"`        // endpoint for s3 storage
	PathPrefix         string            `yaml:"pathPrefix,omitempty"`      // prefix in path for all storage
	Bucket             string            `yaml:"bucket"`
	B2AccountID        string            `yaml:"b2Account"`                    // account name for b2
	B2ApplicationKey   string            `yaml:"b2ApplicationKey"`             // key for b2
	B2ApplicationKeyID string            `yaml:"b2ApplicationKeyId"`           // key for b2
	GoogleConfigJSON   string            `yaml:"googleConfigJson,omitempty"`   // google config json
	GoogleProjectID    string            `yaml:"googleProjectId,omitempty"`    // google project id
	GoogleScopes       string            `yaml:"googleScopes,omitempty"`       // google  scopes id
	OracleUsername     string            `yaml:"oracleUsername,omitempty"`     // oracle user name
	OraclePassword     string            `yaml:"oraclePassword,omitempty"`     // oracle password
	OracleAuthEndpoint string            `yaml:"oracleAuthEndpoint,omitempty"` // oracle auth endpoint
	SFTPHost           string            `yaml:"sftpHost"`                     // host for sftp storage
	SFTPPort           string            `yaml:"sftpPort"`                     // port for sftp storage
	SFTPUsername       string            `yaml:"sftpUsername"`                 // username for sftp storage
	SFTPPassword       string            `yaml:"sftpPassword"`                 // password for sftp storage
	SFTPPrivateKey     string            `yaml:"sftpPrivateKey"`               // private key for sftp
	SFTPPrivateKeyPass string            `yaml:"sftpPrivateKeypassphrase"`     // password for sftp key
	SFTPHostPublicKey  string            `yaml:"sftpHostPublicKey"`            // sft pubic host key
	SFTPHostBasePath   string            `yaml:"sftpBasePath"`                 // base path for sftp storage
	AzureAccount       string            `yaml:"azureAccount,omitempty"`       // azure account name
	AzureKey           string            `yaml:"azureKey,omitempty"`           // azure key
	HTTPTracing        string            `yaml:"HTTPTracing,omitempty" default:"false"`
	Hash               string            // unique hash for given storage
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

func (s *StorageTypes) Noop() Storage {
	return Storage{Kind: "noop"}
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

// GlacierCfg configures S3 GLACIER restore behavior
type GlacierCfg struct {
	Enabled           bool   `yaml:"enabled"`           // Enable auto-restore (default: false - must be explicitly enabled)
	RestoreTier       string `yaml:"restoreTier"`       // Expedited, Standard, Bulk (default: Standard)
	RestoreDays       int    `yaml:"restoreDays"`       // 1-30 days (default: 3)
	RetryAfterSeconds int    `yaml:"retryAfterSeconds"` // Optional, auto-calculated based on tier
}

// Bucket describe single bucket entry in config
type Bucket struct {
	Transform *Transform        `yaml:"transform,omitempty"`
	Storages  StorageTypes      `yaml:"storages"`
	Keys      []S3Key           `yaml:"keys"`
	Headers   map[string]string `yaml:"headers"`
	Glacier   *GlacierCfg       `yaml:"glacier,omitempty"` // GLACIER restore configuration
	Name      string
}

// HeaderYaml allow you to override response headers
type HeaderYaml struct {
	StatusCodes []int             `yaml:"statusCodes"`
	Override    bool              `yaml:"override"`
	Values      map[string]string `yaml:"values"`
}

// CacheCfg configure type of cache
type CacheCfg struct {
	Type             string            `yaml:"type"`
	Address          []string          `yaml:"address"`
	MaxCacheItemSize int64             `yaml:"maxCacheItemSizeMB"`
	CacheSize        int64             `yaml:"cacheSize"`
	MinUseCount      uint64            `yaml:"minUseCount"`
	ClientConfig     map[string]string `yaml:"clientConfig"`
}

// LockCfg configure redis lock
type LockCfg struct {
	Type         string            `yaml:"type"`
	Address      []string          `yaml:"address"`
	ClientConfig map[string]string `yaml:"clientConfig"`
}

// IdleCleanupCfg configures memory cleanup during idle periods
type IdleCleanupCfg struct {
	Enabled        bool `yaml:"enabled"`
	IdleTimeoutMin int  `yaml:"idleTimeoutMin"` // Minutes of inactivity before cleanup
	AggressiveGC   bool `yaml:"aggressiveGC"`   // Run runtime.GC() during cleanup for additional memory recovery
}

// Server configure HTTP server
type Server struct {
	LogLevel                  string                 `yaml:"logLevel"`
	AccessLog                 bool                   `yaml:"accessLogs"`
	InternalListen            string                 `yaml:"internalListen"`
	SingleListen              string                 `yaml:"listen"`
	RequestTimeout            int                    `yaml:"requestTimeout"`
	LockTimeout               int                    `yaml:"lockTimeout"`
	ConcurrentImageProcessing int                    `yaml:"concurrentImageProcessing"` // Maximum number of concurrent image transformations (default: 100)
	Lock                      *LockCfg               `yaml:"lock"`
	Listen                    []string               `yaml:"listens"`
	Monitoring                string                 `yaml:"monitoring"`
	PlaceholderStr            string                 `yaml:"placeholder"`
	Plugins                   map[string]interface{} `yaml:"plugins,omitempty"`
	Cache                     CacheCfg               `yaml:"cache"`
	IdleCleanup               *IdleCleanupCfg        `yaml:"idleCleanup,omitempty"`
	MaxFileSize               int64                  `yaml:"maxFileSize"`
	Placeholder               struct {
		Buf         []byte
		ContentType string
	} `yaml:"-"`
}
