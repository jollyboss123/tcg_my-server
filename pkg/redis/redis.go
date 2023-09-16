package redis

import (
	"context"
	"fmt"
	"github.com/jollyboss123/tcg_my-server/config"
	"github.com/redis/go-redis/v9"
	"log"
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

	err := client.Ping(context.Background()).Err()
	if err != nil {
		log.Fatalf("failed to connect with redis instance at %s - %v", cfg.Host, err)
	}

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
