package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"net/http"

	"github.com/aldor007/mort/pkg/helpers"
	"github.com/aldor007/mort/pkg/monitoring"
)

// Config contains configuration for buckets etc
//
// Config should be used like singleton
type Config struct {
	Buckets         map[string]Bucket `yaml:"buckets"`
	Headers         []HeaderYaml      `yaml:"headers"`
	Server          Server            `yaml:"server"`
	BaseConfigPath  string
	accessKeyBucket map[string][]string
}

var instance *Config
var once sync.Once

// storageKinds is list of available storage kinds
var storageKinds = []string{"local", "local-meta", "s3", "http", "b2", "noop", "google", "azure", "sftp", "oracle"}

// transformKind is list of available kinds of transforms
var transformKinds = []string{"query", "presets", "presets-query", "tengo"}

// GetInstance return single instance of Config object
func GetInstance() *Config {
	once.Do(func() {
		instance = &Config{}
	})
	return instance
}

// RegisterTransformKind register new transformation in config validator
func RegisterTransformKind(kind string) {
	for _, k := range transformKinds {
		if k == kind {
			return
		}
	}

	transformKinds = append(transformKinds, kind)
}

// Load reads config data from file
// How configuration file should be formatted see README.md
func (c *Config) Load(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic("unable to load config file" + err.Error())
	}

	return c.load(data)
}

// LoadFromString parse configuration form string
func (c *Config) LoadFromString(data string) error {
	return c.load([]byte(data))
}

func (c *Config) load(data []byte) error {
	data = []byte(os.ExpandEnv(string(data)))
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
	monitoring.Logs().Warnw(msg)
	return errors.New(msg)
}

func (c *Config) validateStorage(bucketName string, storages StorageTypes) error {
	var validStorageKind bool
	var err error
	basic := storages.Basic()
	if basic.Kind == "" {
		return configInvalidError(fmt.Sprintf("%s basic storage is required", bucketName))
	}

	for storageName, storage := range storages {
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

		if storage.Kind == "s3" || storage.Kind == "s3-fixed" {
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

func (c *Config) validateTransform(bucketName string, bucket *Bucket) error {
	transform := bucket.Transform
	var err error
	errorMsgPrefix := fmt.Sprintf("%s has invalid transform config", bucketName)

	var validTransfromKind bool
	for _, kind := range transformKinds {
		if kind == transform.Kind {
			validTransfromKind = true
			break
		}
	}

	if validTransfromKind == false {
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

	// in case of query string URLs, mort by default generate hash for object
	// example https://mort.mkaciuba.com/demo/img.jpg?operation=rotate&angle=270 will be saved under this path /2c8/img/img.jpg-2c82757531989901
	if transform.ResultKey == "" && (transform.Kind == "query" || transform.Kind == "presets-query") {
		bucket.Transform.ResultKey = "hashParent"
	}

	if transform.Kind == "tengo" {
		buf, errTengo := os.ReadFile(c.getPathToConfig(transform.TengoPath))
		if errTengo != nil {
			err = configInvalidError(fmt.Sprintf("unable to read tengo script file \"%s\", error %v", transform.TengoPath, errTengo))
		}

		t := tengo.NewScript(buf)
		t.SetImports(stdlib.GetModuleMap(stdlib.AllModuleNames()...))
		t.Add("url", nil)
		t.Add("bucketConfig", nil)
		t.Add("obj", nil)
		t.Add("transforms", nil)

		c, errTengo := t.Compile()
		if errTengo != nil {
			err = configInvalidError(fmt.Sprintf("unable to compile tengo script %s error %v", transform.TengoPath, errTengo))
		}
		bucket.Transform.TengoScript = c

	}

	return err

}

func (c *Config) getPathToConfig(filePath string) string {
	if filepath.IsAbs(filePath) {
		return filePath
	}

	basePath := "./configuration/"
	if v := os.Getenv("MORT_CONFIG_DIR"); v != "" {
		basePath = v
	}

	if c.BaseConfigPath != "" {
		basePath = c.BaseConfigPath
	}

	return path.Join(basePath, filePath)
}

func (c *Config) validateServer() error {
	if c.Server.LogLevel == "" {
		c.Server.LogLevel = "prod"
	}

	if c.Server.SingleListen == "" {
		c.Server.SingleListen = ":8080"
	}

	if len(c.Server.Listen) == 0 {
		c.Server.Listen = append(c.Server.Listen, c.Server.SingleListen)
	}

	if c.Server.InternalListen == "" {
		c.Server.InternalListen = ":8081"
	}

	for _, l := range c.Server.Listen {
		if c.Server.InternalListen == l {
			return configInvalidError("Server has invalid configuration internalLstener and listener should have same address")
		}
	}

	if c.Server.Cache.CacheSize == 0 {
		c.Server.Cache.CacheSize = 50000
	}

	if c.Server.RequestTimeout == 0 {
		c.Server.RequestTimeout = 60
	}

	if c.Server.LockTimeout == 0 {
		c.Server.LockTimeout = 30
	}

	if c.Server.Cache.MaxCacheItemSize == 0 {
		c.Server.Cache.MaxCacheItemSize = 5 * 2 << 20
	} else {
		c.Server.Cache.MaxCacheItemSize = c.Server.Cache.MaxCacheItemSize * 2 << 20
	}

	if c.Server.Cache.Type == "" {
		c.Server.Cache.Type = "memory"
	}

	if c.Server.PlaceholderStr != "" {
		buf, err := helpers.FetchObject(c.Server.PlaceholderStr)
		if err != nil {
			return err
		}

		c.Server.Placeholder.Buf = buf
		c.Server.Placeholder.ContentType = http.DetectContentType(buf)
	}

	// Validate idle cleanup configuration
	if c.Server.IdleCleanup != nil && c.Server.IdleCleanup.Enabled {
		if c.Server.IdleCleanup.IdleTimeoutMin == 0 {
			c.Server.IdleCleanup.IdleTimeoutMin = 15 // Conservative default
		}
		if c.Server.IdleCleanup.IdleTimeoutMin < 5 {
			return configInvalidError("idleCleanup.idleTimeoutMin must be at least 5 minutes")
		}
	}

	return nil
}

func (c *Config) validateGlacier(bucketName string, glacier *GlacierCfg) error {
	if glacier == nil {
		return nil
	}

	validTiers := []string{"Expedited", "Standard", "Bulk"}
	tierValid := false
	for _, tier := range validTiers {
		if glacier.RestoreTier == tier {
			tierValid = true
			break
		}
	}

	if glacier.RestoreTier != "" && !tierValid {
		return configInvalidError(fmt.Sprintf("%s has invalid glacier restoreTier '%s', valid: %v",
			bucketName, glacier.RestoreTier, validTiers))
	}

	if glacier.RestoreDays != 0 && (glacier.RestoreDays < 1 || glacier.RestoreDays > 30) {
		return configInvalidError(fmt.Sprintf("%s has invalid glacier restoreDays %d, must be 1-30",
			bucketName, glacier.RestoreDays))
	}

	return nil
}

func (c *Config) validate() error {
	for name, bucket := range c.Buckets {
		err := c.validateStorage(name, bucket.Storages)
		if err != nil {
			return err
		}

		if bucket.Transform != nil {
			err = c.validateTransform(name, &bucket)
			if err != nil {
				return err
			}
		}

		// Validate and set GLACIER defaults
		if bucket.Glacier != nil {
			err = c.validateGlacier(name, bucket.Glacier)
			if err != nil {
				return err
			}

			// Set defaults
			if bucket.Glacier.RestoreTier == "" {
				bucket.Glacier.RestoreTier = "Standard"
			}
			if bucket.Glacier.RestoreDays == 0 {
				bucket.Glacier.RestoreDays = 3
			}
			if bucket.Glacier.RetryAfterSeconds == 0 {
				// Auto-calculate Retry-After based on tier
				switch bucket.Glacier.RestoreTier {
				case "Expedited":
					bucket.Glacier.RetryAfterSeconds = 300 // 5 minutes
				case "Bulk":
					bucket.Glacier.RetryAfterSeconds = 43200 // 12 hours
				default: // Standard
					bucket.Glacier.RetryAfterSeconds = 14400 // 4 hours
				}
			}

			c.Buckets[name] = bucket // Update bucket with defaults
		}
	}
	return c.validateServer()
}
