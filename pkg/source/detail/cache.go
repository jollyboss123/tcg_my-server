package detail

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jollyboss123/tcg_my-server/config"
	"github.com/jollyboss123/tcg_my-server/pkg/game"
	"github.com/jollyboss123/tcg_my-server/pkg/source"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"strings"
)

type cachedDetail struct {
	services map[string]source.DetailService
	gs       game.Service
	cache    *redis.Client
	cfg      *config.Config
	logger   *slog.Logger
}

func NewCachedDetailService(
	cache *redis.Client,
	cfg *config.Config,
	logger *slog.Logger,
	services map[string]source.DetailService,
	gs game.Service,
) source.DetailService {
	child := logger.With(slog.String("api", "cached-detail"))
	return &cachedDetail{
		services: services,
		cache:    cache,
		cfg:      cfg,
		logger:   child,
		gs:       gs,
	}
}

// Fetch fetches cards based on the provided code from the cache or from external services.
// If the cache has entries for the given code, it fetches from the cache; otherwise, it fetches from services.
func (c *cachedDetail) Fetch(ctx context.Context, code, game string) (*source.DetailInfo, error) {
	code = strings.ToUpper(code)
	_, err := c.gs.Fetch(ctx, game)
	if err != nil {
		c.logger.Error("fetch game", slog.String("error", err.Error()), slog.String("code", code), slog.String("game", game))
		return nil, err
	}
	uID := code
	setKey := fmt.Sprintf("game:identifiers:%s", game)
	hashKey := fmt.Sprintf("game:data:%s", game)

	exists, err := c.cache.SIsMember(ctx, setKey, code).Result()
	if err == nil && exists {
		val, err := c.cache.HGet(ctx, hashKey, uID).Result()
		c.logger.Info("cache hit", slog.String("code", code), slog.String("game", game))
		var info *source.DetailInfo
		err = json.Unmarshal([]byte(val), &info)
		if err != nil {
			c.logger.Error("unmarshal cache data", slog.String("error", err.Error()), slog.String("code", code), slog.String("game", game))
			return nil, err
		}
		return info, nil
	}

	c.logger.Info("cache miss", slog.String("code", code), slog.String("game", game))
	s, exists := c.services[game]
	if !exists {
		c.logger.Error("check service exist", slog.String("error", fmt.Errorf("card detail service for %s not found", game).Error()))
		return nil, fmt.Errorf("card detail service for %s not found", game)
	}

	info, err := s.Fetch(ctx, code, game)
	if err != nil {
		c.logger.Error("fetch card detail", slog.String("error", err.Error()), slog.String("game", game), slog.String("code", code))
		return nil, err
	}

	err = c.cache.SAdd(ctx, setKey, uID).Err()
	if err != nil {
		c.logger.Warn("set add cache", slog.String("error", err.Error()), slog.String("code", code), slog.String("game", game))
		return info, nil
	}
	err = c.cache.Expire(ctx, setKey, c.cfg.Cache.CacheTime).Err()
	if err != nil {
		c.logger.Warn("set sets cache expiry", slog.String("error", err.Error()), slog.String("code", code), slog.String("game", game))
		return info, nil
	}

	cacheEntry, err := json.Marshal(info)
	if err != nil {
		c.logger.Warn("marshal cache data", slog.String("error", err.Error()), slog.String("code", code), slog.String("game", game))
		return info, nil
	}
	err = c.cache.HSet(ctx, hashKey, uID, cacheEntry).Err()
	if err != nil {
		c.logger.Warn("set cache", slog.String("error", err.Error()), slog.String("code", code), slog.String("game", game))
		return info, nil
	}
	err = c.cache.Expire(ctx, hashKey, c.cfg.Cache.CacheTime).Err()
	if err != nil {
		c.logger.Warn("set hash set cache expiry", slog.String("error", err.Error()), slog.String("code", code), slog.String("game", game))
		return info, nil
	}

	return info, nil
}
