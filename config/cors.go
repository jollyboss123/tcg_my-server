package config

import "github.com/kelseyhightower/envconfig"

type Cors struct {
	AllowedOrigins      []string `split_words:"true"`
	AllowCredentials    bool     `split_words:"true" default:"true"`
	AllowPrivateNetwork bool     `split_words:"true" default:"false"`
}

func NewCors() Cors {
	var c Cors
	envconfig.MustProcess("CORS", &c)

	return c
}
