package source

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jollyboss123/tcg_my-server/config"
	"github.com/redis/go-redis/v9"
	"log"
)

type CachedSource struct {
	service ScrapeService
	cache   *redis.Client
	cfg     *config.Config
}

func NewCachedScrapeService(service ScrapeService, cache *redis.Client, cfg *config.Config) *CachedSource {
	return &CachedSource{
		service: service,
		cache:   cache,
		cfg:     cfg,
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
			return nil, err
		}

		if len(cards) > 0 {
			return cards, nil
		}
	}

	// cache miss
	cards, err := c.fetchAndCache(ctx, query)
	if err != nil {
		return nil, err
	}

	return cards, nil
}

func (c *CachedSource) fetchAndCache(ctx context.Context, query string) ([]*Card, error) {
	cards, err := c.service.List(ctx, query)
	if err != nil {
		return nil, err
	}
	for _, card := range cards {
		cacheKey := fmt.Sprintf("%s:%s:%s", card.Rarity, card.Name, card.Code)
		cacheEntry, err := json.Marshal(card)
		if err != nil {
			log.Println(err)
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
		log.Println("key", iter.Val())
		var card *Card
		val, err := c.cache.Get(ctx, iter.Val()).Result()
		if err != nil {
			log.Println(err)
			continue
		}
		err = json.Unmarshal([]byte(val), &card)
		if err == nil {
			cards = append(cards, card)
		}
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return cards, nil
}
