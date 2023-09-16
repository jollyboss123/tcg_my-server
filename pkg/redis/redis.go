package redis

import (
	"context"
	"fmt"
	"github.com/jollyboss123/tcg_my-server/config"
	"github.com/redis/go-redis/v9"
	"time"
)

type Cache struct {
	client *redis.Client
	ttl    time.Duration
}

const apqPrefix = "apq:"

func New(cfg config.Cache) *Cache {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Pass,
		DB:       cfg.Name,
	})

	return &Cache{client: client, ttl: cfg.CacheTime}
}

func (c *Cache) Add(ctx context.Context, key string, value interface{}) {
	c.client.Set(ctx, apqPrefix+key, value, c.ttl)
}

func (c *Cache) Get(ctx context.Context, key string) (interface{}, bool) {
	s, err := c.client.Get(ctx, apqPrefix+key).Result()
	if err != nil {
		return struct{}{}, false
	}
	return s, true
}

func (c *Cache) Shutdown(ctx context.Context) *redis.StatusCmd {
	return c.client.Shutdown(ctx)
}
