package redis

import (
	"context"
	"fmt"
	"github.com/jollyboss123/tcg_my-server/config"
	"github.com/redis/go-redis/v9"
	"log/slog"
)

func New(cfg config.Cache, logger *slog.Logger) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Pass,
		DB:       cfg.Name,
	})

	err := client.Ping(context.Background()).Err()
	if err != nil {
		logger.Error("failed to connect with redis instance", slog.String("redis host", cfg.Host),
			slog.String("reason", err.Error()))
	}

	return client
}
