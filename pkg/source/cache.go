package source

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jollyboss123/tcg_my-server/config"
	"github.com/redis/go-redis/v9"
	"log/slog"
)

type CachedSource struct {
	service ScrapeService
	cache   *redis.Client
	cfg     *config.Config
	logger  *slog.Logger
}

func NewCachedScrapeService(service ScrapeService, cache *redis.Client, cfg *config.Config, logger *slog.Logger) *CachedSource {
	child := logger.With(slog.String("api", "cached-scrape"))
	return &CachedSource{
		service: service,
		cache:   cache,
		cfg:     cfg,
		logger:  child,
	}
}

func (c *CachedSource) List(ctx context.Context, query string) ([]*Card, error) {
	var cards []*Card
	patterns := []string{
		fmt.Sprintf("*:%s", query),   // for code
		fmt.Sprintf("*:%s:*", query), // for name
	}

	for _, pattern := range patterns {
		cards, err := c.scan(ctx, pattern)
		if err != nil {
			c.logger.Error("redis scan", slog.String("error", err.Error()), slog.String("pattern", pattern))
			return nil, err
		}

		if len(cards) > 0 {
			c.logger.Info("cache hit", slog.String("query", query), slog.String("pattern", pattern))
			return cards, nil
		}
	}

	c.logger.Info("cache miss", slog.String("query", query))
	cards, err := c.fetchAndCache(ctx, query)
	if err != nil {
		c.logger.Error("fetch and cache", slog.String("error", err.Error()), slog.String("query", query))
		return nil, err
	}

	return cards, nil
}

func (c *CachedSource) fetchAndCache(ctx context.Context, query string) ([]*Card, error) {
	cards, err := c.service.List(ctx, query)
	if err != nil {
		c.logger.Error("list card", slog.String("error", err.Error()), slog.String("query", query))
		return nil, err
	}
	for _, card := range cards {
		cacheKey := fmt.Sprintf("%s:%s:%s", card.Rarity, card.Name, card.Code)
		cacheEntry, err := json.Marshal(card)
		if err != nil {
			c.logger.Warn("cache entry", slog.String("error", err.Error()))
			continue
		}
		err = c.cache.Set(ctx, cacheKey, cacheEntry, c.cfg.Cache.CacheTime).Err()
		if err != nil {
			return c.service.List(ctx, query)
		}
	}
	return cards, nil
}

func (c *CachedSource) scan(ctx context.Context, pattern string) ([]*Card, error) {
	var cards []*Card
	cursor := uint64(0)

	iter := c.cache.Scan(ctx, cursor, pattern, 0).Iterator()
	for iter.Next(ctx) {
		c.logger.Info("key found", slog.String("key", iter.Val()))
		var card *Card
		val, err := c.cache.Get(ctx, iter.Val()).Result()
		if err != nil {
			c.logger.Warn("cache get", slog.String("error", err.Error()), slog.String("key", iter.Val()))
			continue
		}
		err = json.Unmarshal([]byte(val), &card)
		if err == nil {
			cards = append(cards, card)
		}
	}

	if err := iter.Err(); err != nil {
		c.logger.Error("iterate scan", slog.String("error", err.Error()), slog.String("key", iter.Val()))
		return nil, err
	}

	return cards, nil
}
