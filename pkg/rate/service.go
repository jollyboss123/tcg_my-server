package rate

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jollyboss123/tcg_my-server/config"
	"github.com/jollyboss123/tcg_my-server/pkg/currency"
	"log/slog"
	"net/http"
	"net/url"
)

type Service interface {
	Fetch(ctx context.Context, base, to string) (*ExchangeRate, error)
	List(ctx context.Context) ([]*ExchangeRate, error)
}

type ExchangeRate struct {
	From *currency.Currency
	To   *currency.Currency
	Rate float64
}

type service struct {
	endpoint        string
	key             string
	logger          *slog.Logger
	currencyService currency.Service
}

func NewService(logger *slog.Logger, cfg *config.Config, currencyService currency.Service) Service {
	child := logger.With(slog.String("api", "exchange-rate"))
	return &service{
		endpoint:        cfg.Rates.Endpoint,
		key:             cfg.Rates.Key,
		logger:          child,
		currencyService: currencyService,
	}
}

var ErrDataFormat = errors.New("unexpected data format")

func (s *service) Fetch(ctx context.Context, base, to string) (*ExchangeRate, error) {
	return nil, errors.New("fetch from cache")
}

func (s *service) List(ctx context.Context) ([]*ExchangeRate, error) {
	baseURL, err := url.Parse(s.endpoint)
	if err != nil {
		s.logger.Error("parsing url", err.Error(), slog.String("url", s.endpoint))
		return nil, err
	}

	params := url.Values{}
	params.Add("access_key", s.key)
	params.Add("format", "1")
	baseURL.RawQuery = params.Encode()
	r := make([]*ExchangeRate, 0)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL.String(), nil)
	if err != nil {
		s.logger.Error("requesting url", slog.String("error", err.Error()), slog.String("url", baseURL.String()))
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		s.logger.Error("do request", slog.String("error", err.Error()), slog.String("url", baseURL.String()))
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			s.logger.Warn("close response", slog.String("error", cerr.Error()))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("url response", slog.String("error", errors.New("Exchange Rate API returns: "+resp.Status).Error()),
			slog.String("url", baseURL.String()))
		return nil, errors.New("Exchange Rate API returns: " + resp.Status)
	}

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)

	rawRates, ok := data["rates"].(map[string]interface{})
	if !ok {
		s.logger.Error("convert raw json rates", slog.String("error", err.Error()))
		return nil, ErrDataFormat
	}

	base, err := s.currencyService.Fetch(ctx, "EUR")
	if err != nil {
		s.logger.Error("fetch base currency", slog.String("error", err.Error()))
		return nil, err
	}
	for k, v := range rawRates {
		rate, ok := v.(float64)
		if !ok {
			s.logger.Error("convert rate to float64", slog.String("error", err.Error()))
			continue
		}

		dest, err := s.currencyService.Fetch(ctx, k)
		if err != nil {
			s.logger.Error("fetch dest currency", slog.String("error", err.Error()), slog.String("code", k))
			continue
		}
		er := &ExchangeRate{
			From: base,
			To:   dest,
			Rate: rate,
		}
		r = append(r, er)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return r, nil
	}
}
