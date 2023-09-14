package PriceFinder

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type BigWebPriceFinder struct {
	bigwebEndpoint    string
	bigwebAPIEndpoint string
}

func NewBigWebPriceFinder() *BigWebPriceFinder {
	return &BigWebPriceFinder{
		bigwebEndpoint:    "https://bigweb.co.jp/ver2/yugioh_index.php",
		bigwebAPIEndpoint: "https://api.bigweb.co.jp/products",
	}
}

func (f *BigWebPriceFinder) FindPrices(jpName string) {
	jpName = formatJapaneseName(jpName)

	params := url.Values{}
	params.Add("game_id", "9")
	params.Add("name", jpName)

	resp, err := http.Get(f.bigwebAPIEndpoint + "?" + params.Encode())
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

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

		//cardID := cardInfo["fname"].(string)
		rawRarity := cardInfo["rarity"].(map[string]interface{})["slip"].(string)
		rarity := formatRarity(rawRarity)
		rawCondition := cardInfo["condition"].(map[string]interface{})["slip"].(string)
		condition := "Scratch"
		if rawCondition != "キズ" {
			condition = "Play"
		}
		price := cardInfo["price"].(float64)

		fmt.Printf("name: %s rarity: %s condition: %s price: %.2f\n", jpName, rarity, condition, price)
	}
}

func formatJapaneseName(name string) string {
	name = strings.Replace(name, "－", " ", -1)
	return name
}

func formatRarity(rarity string) string {
	// Implement rarity formatting logic based on your requirements
	return rarity
}
