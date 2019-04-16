package cache

import (
	"github.com/aldor007/mort/pkg/config"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestCreateDefault(t *testing.T) {
	cfg := config.CacheCfg{}
	instance := Create(cfg)

	assert.Equal(t, reflect.TypeOf(instance).String(), reflect.TypeOf(&MemoryCache{}).String())
}

func TestCreatRedis(t *testing.T) {
	cfg := config.CacheCfg{}
	cfg.Type = "redis"
	instance := Create(cfg)

	assert.Equal(t, reflect.TypeOf(instance).String(), reflect.TypeOf(&RedisCache{}).String())
}
