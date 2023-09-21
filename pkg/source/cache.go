package source

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jollyboss123/tcg_my-server/config"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"strings"
	"sync"
)

type CachedSource struct {
	services []ScrapeService
	cache    *redis.Client
	cfg      *config.Config
	logger   *slog.Logger
}

func NewCachedScrapeService(cache *redis.Client, cfg *config.Config, logger *slog.Logger, service ...ScrapeService) *CachedSource {
	child := logger.With(slog.String("api", "cached-scrape"))
	return &CachedSource{
		services: service,
		cache:    cache,
		cfg:      cfg,
		logger:   child,
	}
}

func (c *CachedSource) List(ctx context.Context, query string) ([]*Card, error) {
	query = strings.ToUpper(query)

	if c.isQueryCached(ctx, query) {
		return c.fetchFromDataCache(ctx, query)
	}

	cards, err := c.fetchAndCache(ctx, query)
	if err != nil {
		return nil, err
	}

	return cards, nil
}

func (c *CachedSource) isQueryCached(ctx context.Context, query string) bool {
	exists, err := c.cache.Exists(ctx, fmt.Sprintf("query:%s", query)).Result()
	return err == nil && exists > 0
}

func (c *CachedSource) cacheQuery(ctx context.Context, query string) error {
	err := c.cache.Set(ctx, fmt.Sprintf("query:%s", query), "true", c.cfg.Cache.CacheTime).Err()
	if err != nil {
		return err
	}
	return nil
}

func (c *CachedSource) fetchFromDataCache(ctx context.Context, query string) ([]*Card, error) {
	var cards []*Card
	var mu sync.Mutex

	patterns := []string{
		fmt.Sprintf("*:%s", query),   // for code
		fmt.Sprintf("*:%s:*", query), // for name
		fmt.Sprintf("*:%s-*", query), // for booster pack
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(patterns))

	errCh := make(chan error, len(patterns))

	for _, pattern := range patterns {
		go func(p string) {
			defer wg.Done()

			cs, err := c.scan(ctx, p)
			if err != nil {
				c.logger.Error("redis scan", slog.String("error", err.Error()), slog.String("pattern", p))
				errCh <- err
				return
			}

			if len(cs) <= 0 {
				return
			}

			c.logger.Info("cache hit", slog.String("query", query), slog.String("pattern", p))
			mu.Lock()
			cards = append(cards, cs...)
			mu.Unlock()
		}(pattern)
	}

	wg.Wait()
	close(errCh)

	if len(errCh) > 0 {
		return nil, <-errCh
	}

	if len(cards) > 0 {
		return cards, nil
	}

	return cards, nil
}

func (c *CachedSource) fetchAndCache(ctx context.Context, query string) ([]*Card, error) {
	var cards []*Card
	var mu sync.Mutex

	wg := &sync.WaitGroup{}
	wg.Add(len(c.services))

	for _, service := range c.services {
		go func(s ScrapeService) {
			defer wg.Done()

			cs, err := s.List(ctx, query)
			if err != nil {
				c.logger.Error("list card", slog.String("error", err.Error()), slog.String("query", query))
				return
			}

			mu.Lock()
			cards = append(cards, cs...)
			mu.Unlock()
		}(service)
	}

	wg.Wait()

	if len(cards) == 0 {
		c.logger.Info("no cards found", slog.String("query", query))
		return cards, nil
	}

	for _, card := range cards {
		cacheKey := fmt.Sprintf("%s:%s:%s", card.Rarity, card.Name, card.Code)
		_ = c.cacheQuery(ctx, card.Name)
		_ = c.cacheQuery(ctx, card.Code)
		cacheEntry, err := json.Marshal(card)
		if err != nil {
			c.logger.Warn("cache entry", slog.String("error", err.Error()), slog.String("query", query))
			continue
		}
		err = c.cache.Set(ctx, cacheKey, cacheEntry, c.cfg.Cache.CacheTime).Err()
		if err != nil {
			c.logger.Warn("set cache", slog.String("error", err.Error()), slog.String("query", query))
			continue
		}
	}
	return cards, nil
}

func (c *CachedSource) scan(ctx context.Context, pattern string) ([]*Card, error) {
	var cards []*Card
	cursor := uint64(0)

	iter := c.cache.Scan(ctx, cursor, pattern, 0).Iterator()
	for iter.Next(ctx) {
		c.logger.Debug("key found", slog.String("key", iter.Val()))
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
