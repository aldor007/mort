package cache

import (
	"reflect"
	"testing"

	"github.com/aldor007/mort/pkg/config"
	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func TestCreateDefault(t *testing.T) {
	cfg := config.CacheCfg{}
	instance := Create(cfg)

	assert.Equal(t, reflect.TypeOf(instance).String(), reflect.TypeOf(&MemoryCache{}).String())
}

func TestCreatRedis(t *testing.T) {
	s := miniredis.RunT(t)

	cfg := config.CacheCfg{}
	cfg.Type = "redis"
	cfg.Address = []string{s.Addr()}
	instance := Create(cfg)

	assert.Equal(t, reflect.TypeOf(instance).String(), reflect.TypeOf(&RedisCache{}).String())
}
