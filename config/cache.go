package config

import (
	"github.com/kelseyhightower/envconfig"
	"strings"
	"time"
)

type Cache struct {
	Enable    bool   `default:"false"`
	Host      string `default:"0.0.0.0"`
	Port      string `default:"6379"`
	Hosts     []string
	Name      int `default:"1"`
	User      string
	Pass      string
	CacheTime time.Duration `split_words:"true" default:"5m"`
}

func NewCache() Cache {
	var cache Cache
	envconfig.MustProcess("REDIS", &cache)

	if strings.Contains("Host", ",") {
		cache.Hosts = strings.Split(cache.Host, ",")
	}

	return cache
}
