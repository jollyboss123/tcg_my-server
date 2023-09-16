package source

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

func (f *BigWeb) List(ctx context.Context, code string) ([]*Card, error) {
	params := url.Values{}
	params.Add("game_id", "9")
	params.Add("Name", code)
	u := f.endpoint + "?" + params.Encode()
	c := make([]*Card, 0)

	resp, err := http.Get(u)
	if err != nil {
		fmt.Println("Error:", err)
		return c, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error:", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return c, errors.New("BigWeb API returns: " + string(rune(resp.StatusCode)))
	}

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return c, err
	}

	rawCardInfos := data["items"].([]interface{})
	for _, rawCardInfo := range rawCardInfos {
		card := Card{}
		info := rawCardInfo.(map[string]interface{})

		card.Code = info["fname"].(string)
		card.Rarity = info["Rarity"].(map[string]interface{})["slip"].(string)
		rawCondition := info["Condition"].(map[string]interface{})["slip"].(string)
		card.Condition = "Scratch"
		if rawCondition != "キズ" {
			card.Condition = "Play"
		}
		card.Price = int64(info["Price"].(float64))

		c = append(c, &card)

		fmt.Printf("Name: %s Rarity: %s Condition: %s Price: %d\n", card.Code, card.Rarity, card.Condition, card.Price)
	}

	return c, nil
}
