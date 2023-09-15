package source

import (
	"errors"
	"fmt"
	"github.com/gocolly/colly/v2"
	"net/url"
	"strconv"
	"time"
	"unicode"
)

type YYT struct {
	endpoint string
	source   string
}

func NewYYT() *YYT {
	return &YYT{
		endpoint: "https://yuyu-tei.jp/sell/ygo/s/search",
		source:   "Yuyu-tei",
	}
}

func (y *YYT) Scrape(code string) (CardInfo, error) {
	cardInfo := CardInfo{url: y.endpoint + "?search_word=" + url.QueryEscape(code)}

	c := colly.NewCollector(
		colly.AllowedDomains("yuyu-tei.jp"),
		colly.AllowURLRevisit(),
		colly.Async(true),
	)

	err := c.Limit(&colly.LimitRule{
		DomainGlob:  "yuyu-tei.jp/*",
		Delay:       1 * time.Second,
		RandomDelay: 1 * time.Second,
	})
	if err != nil {
		return CardInfo{}, err
	}

	ch := make(chan error)
	defer func() { ch <- nil }()

	c.OnHTML("div[code=card-list3]", func(e *colly.HTMLElement) {
		rarity := e.ChildText("h3 > span")
		e.ForEach("div[code=card-lits]", func(_ int, el *colly.HTMLElement) {
			card := Card{}

			card.code = substring(e.ChildText("span"), 2)
			card.price, err = strconv.ParseInt(extractNumbers(e.ChildText("strong")), 16, 64)
			if err != nil {
				ch <- err
			}
			card.rarity = rarity

			cardInfo.cards = append(cardInfo.cards, card)
			fmt.Printf("name: %s rarity: %s price: %d\n", card.code, card.rarity, card.price)
		})
	})

	numVisited := 0
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("visiting", r.URL.String())
		if numVisited > 100 {
			r.Abort()
			ch <- errors.New("request limit reached")
		}
		numVisited++
	})

	err = c.Visit(cardInfo.url)
	if err != nil {
		return CardInfo{}, err
	}

	result := <-ch
	if result != nil {
		return CardInfo{}, result
	}

	return cardInfo, nil
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
