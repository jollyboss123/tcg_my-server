package redis

import (
	"context"
	"fmt"
	"github.com/jollyboss123/tcg_my-server/config"
	"github.com/redis/go-redis/v9"
	"log"
)

func New(cfg config.Cache) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Pass,
		DB:       cfg.Name,
	})

	err := client.Ping(context.Background()).Err()
	if err != nil {
		log.Fatalf("failed to connect with redis instance at %s - %v", cfg.Host, err)
	}

	return client
}
