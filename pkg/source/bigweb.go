package source

import (
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

func (f *BigWeb) Scrape(code string) (CardInfo, error) {
	params := url.Values{}
	params.Add("game_id", "9")
	params.Add("name", code)
	cardInfo := CardInfo{url: f.endpoint + "?" + params.Encode()}

	resp, err := http.Get(cardInfo.url)
	if err != nil {
		fmt.Println("Error:", err)
		return CardInfo{}, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error:", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return CardInfo{}, errors.New("BigWeb API returns: " + string(rune(resp.StatusCode)))
	}

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return CardInfo{}, err
	}

	rawCardInfos := data["items"].([]interface{})
	for _, rawCardInfo := range rawCardInfos {
		card := Card{}
		info := rawCardInfo.(map[string]interface{})

		card.code = info["fname"].(string)
		card.rarity = info["rarity"].(map[string]interface{})["slip"].(string)
		rawCondition := info["condition"].(map[string]interface{})["slip"].(string)
		card.condition = "Scratch"
		if rawCondition != "キズ" {
			card.condition = "Play"
		}
		card.price = int64(info["price"].(float64))

		cardInfo.cards = append(cardInfo.cards, card)

		fmt.Printf("name: %s rarity: %s condition: %s price: %d\n", card.code, card.rarity, card.condition, card.price)
	}

	return cardInfo, nil
}
