package rate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jollyboss123/tcg_my-server/config"
	"github.com/jollyboss123/tcg_my-server/pkg/currency"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"strings"
)

type CachedExchangeRate struct {
	service         Service
	cache           *redis.Client
	cfg             *config.Config
	logger          *slog.Logger
	currencyService currency.Service
}

func NewCachedExchangeRate(service Service, cache *redis.Client, cfg *config.Config, logger *slog.Logger, currencyService currency.Service) *CachedExchangeRate {
	child := logger.With(slog.String("api", "cached-exchange-rate"))
	return &CachedExchangeRate{
		service:         service,
		cache:           cache,
		cfg:             cfg,
		logger:          child,
		currencyService: currencyService,
	}
}

func (c *CachedExchangeRate) Fetch(ctx context.Context, base, to string) (*ExchangeRate, error) {
	base = strings.ToUpper(base)
	to = strings.ToUpper(to)
	cacheMiss := false
	var br *ExchangeRate
	var dr *ExchangeRate

	_, err := c.currencyService.Fetch(ctx, base)
	if errors.Is(err, currency.ErrCurrencyNotFound) {
		c.logger.Error("check base currency exist", slog.String("error", err.Error()), slog.String("base", base), slog.String("to", to))
		return nil, err
	}

	_, err = c.currencyService.Fetch(ctx, to)
	if errors.Is(err, currency.ErrCurrencyNotFound) {
		c.logger.Error("check dest currency exist", slog.String("error", err.Error()), slog.String("base", base), slog.String("to", to))
		return nil, err
	}

	b, err := c.cache.Get(ctx, "rate:EUR:"+base).Result()
	if err != nil {
		c.logger.Error("cache miss for base", slog.String("error", err.Error()), slog.String("base", base), slog.String("to", to))
		cacheMiss = true
	}

	d, err := c.cache.Get(ctx, "rate:EUR:"+to).Result()
	if err != nil {
		c.logger.Error("cache miss for dest", slog.String("error", err.Error()), slog.String("base", base), slog.String("to", to))
		cacheMiss = true
	}

	if cacheMiss != true {
		err = json.Unmarshal([]byte(b), &br)
		if err != nil {
			c.logger.Error("unmarshal base exchange rate", slog.String("error", err.Error()), slog.String("base", base), slog.String("to", to))
			return nil, err
		}
		err = json.Unmarshal([]byte(d), &dr)
		if err != nil {
			c.logger.Error("unmarshal dest exchange rate", slog.String("error", err.Error()), slog.String("base", base), slog.String("to", to))
			return nil, err
		}
	} else {
		cache, err := c.fetchAndCache(ctx)
		if err != nil {
			c.logger.Error("fetch and cache", slog.String("error", err.Error()), slog.String("base", base), slog.String("to", to))
			return nil, err
		}
		for _, er := range cache {
			if base == er.To.Code {
				br = er
			}
			if to == er.To.Code {
				dr = er
			}
			if br != nil && er != nil {
				break
			}
		}
		if br == nil || dr == nil {
			c.logger.Error("Could not find exchange rates", slog.String("base", base), slog.String("to", to))
			return nil, fmt.Errorf("could not find exchange rates for base: %s and/or to: %s", base, to)
		}
	}

	return &ExchangeRate{
		From: br.To,
		To:   dr.To,
		Rate: br.Rate / dr.Rate,
	}, nil
}

func (c *CachedExchangeRate) List(ctx context.Context) ([]*ExchangeRate, error) {
	val, err := c.cache.Get(ctx, "rates").Result()
	if err != nil {
		return c.fetchAndCache(ctx)
	}
	var keys []string
	err = json.Unmarshal([]byte(val), &keys)
	if err != nil {
		c.logger.Error("unmarshal keys", slog.String("error", err.Error()))
		return nil, err
	}
	var res []*ExchangeRate
	for _, key := range keys {
		val, err := c.cache.Get(ctx, key).Result()
		if err != nil {
			c.logger.Warn("cache miss", slog.String("error", err.Error()), slog.String("key", key))
			continue
		}
		var rate *ExchangeRate
		err = json.Unmarshal([]byte(val), &rate)
		if err != nil {
			c.logger.Warn("unmarshal rate", slog.String("error", err.Error()), slog.String("key", key))
			continue
		}
		res = append(res, rate)
	}
	return res, nil
}

func (c *CachedExchangeRate) fetchAndCache(ctx context.Context) ([]*ExchangeRate, error) {
	rates, err := c.service.List(ctx)
	if err != nil {
		c.logger.Error("list rates", slog.String("error", err.Error()))
		return nil, err
	}
	if len(rates) == 0 {
		return make([]*ExchangeRate, 0), nil
	}

	var keys []string
	for _, rate := range rates {
		key := "rate:" + rate.From.Code + ":" + rate.To.Code
		cacheEntry, err := json.Marshal(rate)
		if err != nil {
			c.logger.Error("cache entry", slog.String("error", err.Error()))
			continue
		}
		err = c.cache.Set(ctx, key, cacheEntry, c.cfg.Cache.CacheTime).Err()
		if err != nil {
			c.logger.Warn("set keys cache", slog.String("error", err.Error()))
			continue
		}
		keys = append(keys, key)
	}
	cacheEntry, err := json.Marshal(keys)
	if err != nil {
		c.logger.Error("cache entry", slog.String("error", err.Error()))
		return rates, nil
	}
	err = c.cache.Set(ctx, "rates", cacheEntry, c.cfg.Cache.CacheTime).Err()
	if err != nil {
		c.logger.Error("set rates cache", slog.String("error", err.Error()))
		return rates, nil
	}
	return rates, nil
}
