package config

import (
	"github.com/joho/godotenv"
	"log"
)

type Config struct {
	Api
	Cache
	Cors
}

func New() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println(err)
	}

	return &Config{
		Api:   API(),
		Cache: NewCache(),
		Cors:  NewCors(),
	}
}
