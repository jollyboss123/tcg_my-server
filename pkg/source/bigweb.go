package source

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
)

type BigWeb struct {
	endpoint string
	source   string
	logger   *slog.Logger
}

func NewBigWeb(logger *slog.Logger) *BigWeb {
	child := logger.With(slog.String("api", "bigweb"))
	return &BigWeb{
		endpoint: "https://api.bigweb.co.jp/products",
		source:   "bigweb",
		logger:   child,
	}
}

var (
	ErrDataFormat     = errors.New("unexpected data format")
	ErrCardDataFormat = errors.New("unexpected card data format")
)

func (b *BigWeb) List(ctx context.Context, query string) ([]*Card, error) {
	baseURL, err := url.Parse(b.endpoint)
	if err != nil {
		b.logger.Error("parsing url", err.Error(), slog.String("url", b.endpoint))
		return nil, err
	}

	params := url.Values{}
	params.Add("game_id", "9")
	params.Add("Name", query)
	baseURL.RawQuery = params.Encode()
	c := make([]*Card, 0)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL.String(), nil)
	if err != nil {
		b.logger.Error("requesting url", slog.String("error", err.Error()), slog.String("url", baseURL.String()))
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		b.logger.Error("do request", slog.String("error", err.Error()), slog.String("url", baseURL.String()))
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			b.logger.Warn("close response", slog.String("error", cerr.Error()))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		b.logger.Error("url response", slog.String("error", errors.New("BigWeb API returns: "+resp.Status).Error()),
			slog.String("url", baseURL.String()))
		return nil, errors.New("BigWeb API returns: " + resp.Status)
	}

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		b.logger.Error("decoding json", slog.String("error", err.Error()), slog.String("url", baseURL.String()))
		return nil, err
	}

	rawCardInfos, ok := data["items"].([]interface{})
	if !ok {
		b.logger.Error("convert raw json items", slog.String("error", err.Error()))
		return nil, ErrDataFormat
	}

	for _, rawCardInfo := range rawCardInfos {
		card := Card{}
		info, ok := rawCardInfo.(map[string]interface{})
		if !ok {
			b.logger.Error("convert card info", slog.String("error", err.Error()))
			return nil, ErrCardDataFormat
		}

		card.Code, _ = info["fname"].(string)
		rarityMap, ok := info["Rarity"].(map[string]interface{})
		if ok {
			card.Rarity, _ = rarityMap["slip"].(string)
		}
		conditionMap, ok := info["Condition"].(map[string]interface{})
		if ok {
			rawCondition, _ := conditionMap["slip"].(string)
			card.Condition = "Scratch"
			if rawCondition != "キズ" {
				card.Condition = "Play"
			}
		}
		priceFloat, ok := info["Price"].(float64)
		if ok {
			card.Price = int64(priceFloat)
		}
		card.Source = b.source

		c = append(c, &card)

		b.logger.Info("card info", slog.String("name", card.Name),
			slog.String("code", card.Code),
			slog.String("rarity", card.Rarity),
			slog.String("condition", card.Condition),
			slog.Int64("price", card.Price))
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return c, nil
	}
}
