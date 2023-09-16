package source

import (
	"context"
	"encoding/json"
	redisLib "github.com/jollyboss123/tcg_my-server/pkg/redis"
)

type CachedSource struct {
	service ScrapeService
	cache   *redisLib.Cache
}

type CachedScrapeService interface {
	List(ctx context.Context, code string) ([]*Card, error)
}

func NewCachedScrapeService(service ScrapeService, cache *redisLib.Cache) *CachedSource {
	return &CachedSource{
		service: service,
		cache:   cache,
	}
}

func (c *CachedSource) List(ctx context.Context, code string) ([]*Card, error) {
	val, ok := c.cache.Get(ctx, code)
	if !ok {
		cards, err := c.service.List(ctx, code)
		if err != nil {
			return nil, err
		}
		cacheEntry, err := json.Marshal(cards)
		if err != nil {
			return c.service.List(ctx, code)
		}

		c.cache.Add(ctx, code, cacheEntry)

		return cards, nil
	}

	var res []*Card
	err := json.Unmarshal([]byte(val.(string)), &res)
	if err != nil {
		return c.service.List(ctx, code)
	}
	return res, nil
}
