package config

import (
	"github.com/kelseyhightower/envconfig"
	"time"
)

type Rates struct {
	JobID    string        `split_words:"true" default:"fetch.exchange-rates"`
	Enabled  bool          `split_words:"true" default:"true"`
	Start    string        `split_words:"true"`
	Delay    time.Duration `split_words:"true" default:"24h"`
	Endpoint string
	Key      string
}

func NewRates() Rates {
	var rates Rates
	envconfig.MustProcess("RATE", &rates)

	return rates
}
