package config

import (
	"github.com/joho/godotenv"
	"log/slog"
	"os"
)

type Config struct {
	Api
	Cache
	Cors
	Rates
}

func New(logger *slog.Logger) *Config {
	if os.Getenv("APP_ENV") != "prod" {
		err := godotenv.Load()
		if err != nil {
			logger.Error("build config", slog.String("error", err.Error()))
		}
	}

	return &Config{
		Api:   API(),
		Cache: NewCache(),
		Cors:  NewCors(),
		Rates: NewRates(),
	}
}
