package pricefinder

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type BigwebPriceFinder struct {
	bigwebEndpoint    string
	bigwebAPIEndpoint string
}

func NewBigWebPriceFinder() *BigwebPriceFinder {
	return &BigwebPriceFinder{
		bigwebEndpoint:    "https://bigweb.co.jp/ver2/yugioh_index.php",
		bigwebAPIEndpoint: "https://api.bigweb.co.jp/products",
	}
}

func (f *BigwebPriceFinder) FindPrices(jpName string) {
	params := url.Values{}
	params.Add("game_id", "9")
	params.Add("name", jpName)

	resp, err := http.Get(f.bigwebAPIEndpoint + "?" + params.Encode())
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error:", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Println("BigWeb API returns", resp.StatusCode)
		return
	}

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	rawCardInfos := data["items"].([]interface{})
	for _, rawCardInfo := range rawCardInfos {
		cardInfo := rawCardInfo.(map[string]interface{})

		cardID := cardInfo["fname"].(string)
		rawRarity := cardInfo["rarity"].(map[string]interface{})["slip"].(string)
		rarity := rawRarity
		rawCondition := cardInfo["condition"].(map[string]interface{})["slip"].(string)
		condition := "Scratch"
		if rawCondition != "キズ" {
			condition = "Play"
		}
		price := cardInfo["price"].(float64)

		fmt.Printf("name: %s rarity: %s condition: %s price: %.2f\n", cardID, rarity, condition, price)
	}
}
