package source

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jollyboss123/tcg_my-server/config"
	"github.com/jollyboss123/tcg_my-server/pkg/game"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"strings"
	"sync"
)

type CachedSource struct {
	services []ScrapeService
	gs       game.Service
	cache    *redis.Client
	cfg      *config.Config
	logger   *slog.Logger
}

func NewCachedScrapeService(cache *redis.Client, cfg *config.Config, logger *slog.Logger, gs game.Service, service ...ScrapeService) *CachedSource {
	child := logger.With(slog.String("api", "cached-scrape"))
	return &CachedSource{
		services: service,
		cache:    cache,
		cfg:      cfg,
		logger:   child,
		gs:       gs,
	}
}

// List fetches cards based on the provided query from the cache or from external services.
// If the cache has entries for the given query, it fetches from the cache; otherwise, it fetches from services.
func (c *CachedSource) List(ctx context.Context, query, game string) ([]*Card, error) {
	query = strings.ToUpper(query)
	_, err := c.gs.Fetch(ctx, game)
	if err != nil {
		c.logger.Error("fetch game", slog.String("error", err.Error()), slog.String("query", query), slog.String("game", game))
		return make([]*Card, 0), err
	}

	if c.isQueryCached(ctx, query, game) {
		c.logger.Info("cache hit", slog.String("query", query))
		return c.fetchFromDataCache(ctx, query, game)
	}

	c.logger.Info("cache miss", slog.String("query", query))
	cards, err := c.fetchAndCache(ctx, query, game)
	if err != nil {
		return nil, err
	}

	return cards, nil
}

func (c *CachedSource) isQueryCached(ctx context.Context, query, game string) bool {
	exists, err := c.cache.SIsMember(ctx, fmt.Sprintf("query:%s", game), query).Result()
	return err == nil && exists
}

func (c *CachedSource) cacheQuery(ctx context.Context, query, game string) error {
	err := c.cache.SAdd(ctx, fmt.Sprintf("query:%s", game), query).Err()
	if err != nil {
		return err
	}
	return nil
}

func (c *CachedSource) fetchFromDataCache(ctx context.Context, query, game string) ([]*Card, error) {
	var cards []*Card
	var mu sync.Mutex
	setKey := fmt.Sprintf("game:identifiers:%s", game)
	hashKey := fmt.Sprintf("game:data:%s", game)

	patterns := []string{
		fmt.Sprintf("*||*||*%s*", query), // for code or booster
		fmt.Sprintf("*||*%s*||*", query), // for name
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(patterns))

	errCh := make(chan error, len(patterns))

	for _, pattern := range patterns {
		go func(p string) {
			defer wg.Done()

			cs, err := c.sscan(ctx, setKey, hashKey, p)
			if err != nil {
				c.logger.Error("redis scan", slog.String("error", err.Error()), slog.String("pattern", p))
				errCh <- err
				return
			}

			if len(cs) <= 0 {
				return
			}

			c.logger.Info("cache hit", slog.String("query", query), slog.String("pattern", p), slog.Int("total", len(cs)))
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

func (c *CachedSource) fetchAndCache(ctx context.Context, query, game string) ([]*Card, error) {
	var cards []*Card
	var mu sync.Mutex

	wg := &sync.WaitGroup{}
	wg.Add(len(c.services))

	for _, service := range c.services {
		go func(s ScrapeService) {
			defer wg.Done()

			cs, err := s.List(ctx, query, game)
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

	// Assign scores based on rarity. Assuming the list is sorted such that
	// the rarest card is first and the least rare card is last.
	for idx, card := range cards {
		card.Score = len(cards) - idx
		_ = c.cacheQuery(ctx, card.Name, game)
		_ = c.cacheQuery(ctx, card.Code, game)

		uID := fmt.Sprintf("%s||%s||%s", card.Rarity, card.Name, card.Code)
		setKey := fmt.Sprintf("game:identifiers:%s", game)
		hashKey := fmt.Sprintf("game:data:%s", game)

		err := c.cache.SAdd(ctx, setKey, uID).Err()
		if err != nil {
			c.logger.Warn("set add cache", slog.String("error", err.Error()), slog.String("query", query))
			continue
		}
		data, err := json.Marshal(card)
		if err != nil {
			c.logger.Warn("marshal cache data", slog.String("error", err.Error()), slog.String("query", query))
			continue
		}
		err = c.cache.HSet(ctx, hashKey, uID, data).Err()
		if err != nil {
			c.logger.Warn("cache entry", slog.String("error", err.Error()), slog.String("query", query))
			continue
		}
	}
	c.logger.Info("cache entry", slog.String("query", query), slog.Int("total", len(cards)))
	_ = c.cacheQuery(ctx, query, game)
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

func (c *CachedSource) sscan(ctx context.Context, setKey, hashKey, pattern string) ([]*Card, error) {
	var cards []*Card
	cursor := uint64(0)

	iter := c.cache.SScan(ctx, setKey, cursor, pattern, 0).Iterator()
	for iter.Next(ctx) {
		c.logger.Debug("key found", slog.String("key", iter.Val()))
		var card *Card
		val, err := c.cache.HGet(ctx, hashKey, iter.Val()).Result()
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
