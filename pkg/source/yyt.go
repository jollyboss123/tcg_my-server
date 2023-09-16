package source

import (
	"context"
	"errors"
	"fmt"
	"github.com/gocolly/colly/v2"
	"net/url"
	"strconv"
	"sync"
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

func (y *YYT) List(ctx context.Context, code string) ([]*Card, error) {
	u := y.endpoint + "?search_word=" + url.QueryEscape(code)
	cs := make([]*Card, 0)

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
		return cs, err
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 1)

	c.OnHTML("div[id=card-list3]", func(e *colly.HTMLElement) {
		rarity := e.ChildText("h3 > span")
		e.ForEach("div[id=card-lits]", func(_ int, el *colly.HTMLElement) {
			card := Card{}

			card.Code = substring(e.ChildText("span"), 2)
			card.Price, err = strconv.ParseInt(extractNumbers(e.ChildText("strong")), 10, 64)
			if err != nil {
				errCh <- err
			}
			card.Rarity = rarity

			cs = append(cs, &card)
			fmt.Printf("Name: %s Rarity: %s Price: %d\n", card.Code, card.Rarity, card.Price)
		})
	})

	numVisited := 0
	c.OnRequest(func(r *colly.Request) {
		wg.Add(1)
		fmt.Println("visiting", r.URL.String())
		if numVisited > 100 {
			r.Abort()
			errCh <- errors.New("request limit reached")
		}
		numVisited++
	})

	c.OnResponse(func(_ *colly.Response) {
		wg.Done()
	})

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := c.Visit(u); err != nil {
			errCh <- err
		}
	}()

	go func() {
		wg.Wait()
		close(errCh)
	}()

	if err := <-errCh; err != nil {
		return nil, err
	}

	c.Wait()
	return cs, nil
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
