package source

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

type BigWeb struct {
	endpoint string
	source   string
}

func NewBigWeb() *BigWeb {
	return &BigWeb{
		endpoint: "https://api.bigweb.co.jp/products",
		source:   "bigweb",
	}
}

func (f *BigWeb) List(ctx context.Context, query string) ([]*Card, error) {
	baseURL, err := url.Parse(f.endpoint)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Add("game_id", "9")
	params.Add("Name", query)
	baseURL.RawQuery = params.Encode()
	c := make([]*Card, 0)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			fmt.Println("Error:", cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("BigWeb API returns: " + resp.Status)
	}

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	rawCardInfos, ok := data["items"].([]interface{})
	if !ok {
		return nil, errors.New("unexpected data format")
	}

	for _, rawCardInfo := range rawCardInfos {
		card := Card{}
		info, ok := rawCardInfo.(map[string]interface{})
		if !ok {
			return nil, errors.New("unexpected card data format")
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
		card.Source = f.source

		c = append(c, &card)

		fmt.Printf("Name: %s Rarity: %s Condition: %s Price: %d\n", card.Code, card.Rarity, card.Condition, card.Price)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return c, nil
	}
}
