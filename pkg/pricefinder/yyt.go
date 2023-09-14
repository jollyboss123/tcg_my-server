package pricefinder

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"log"
	"net/url"
	"unicode"
)

type YuyuteiPriceFinder struct {
	yuyuteiEndpoint string
	source          string
	yuyuteiIcon     string
}

type Card struct {
	cardID    string
	jpName    string
	rarity    string
	condition string
	price     string
}

type CardInfo struct {
	url   string
	cards []Card
}

func NewYuyuteiPriceFinder() *YuyuteiPriceFinder {
	return &YuyuteiPriceFinder{
		yuyuteiEndpoint: "https://yuyu-tei.jp/sell/ygo/s/search",
		source:          "Yuyu-tei",
		yuyuteiIcon:     "https://yuyu-tei.jp/img/ogp.jpg",
	}
}

func (y *YuyuteiPriceFinder) FindPrices(jpName string) CardInfo {
	cardInfo := CardInfo{url: y.yuyuteiEndpoint + "?search_word=" + url.QueryEscape(jpName)}
	fmt.Println(cardInfo.url)
	c := colly.NewCollector()

	c.OnHTML("div[id=card-list3]", func(e *colly.HTMLElement) {
		rarity := e.ChildText("h3 > span")
		e.ForEach("div[id=card-lits]", func(_ int, el *colly.HTMLElement) {
			card := Card{}

			card.cardID = substring(e.ChildText("span"), 2)
			card.price = extractNumbers(e.ChildText("strong"))
			card.rarity = rarity

			cardInfo.cards = append(cardInfo.cards, card)
			fmt.Printf("name: %s rarity: %s price: %s\n", card.cardID, card.rarity, card.price)
		})
	})

	err := c.Visit(cardInfo.url)
	if err != nil {
		log.Fatal(err)
	}

	return cardInfo
}

func extractNumbers(s string) string {
	var result string
	for _, r := range s {
		if unicode.IsDigit(r) {
			result += string(r)
		}
	}
	return result
}

func substring(s string, from int) string {
	return s[from:]
}
