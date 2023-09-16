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
	baseURL, err := url.Parse(y.endpoint)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Add("search_word", code)
	baseURL.RawQuery = params.Encode()
	cs := make([]*Card, 0)

	c := colly.NewCollector(
		colly.AllowedDomains("yuyu-tei.jp"),
		colly.AllowURLRevisit(),
		colly.Async(true),
	)

	err = c.Limit(&colly.LimitRule{
		DomainGlob:  "yuyu-tei.jp/*",
		Delay:       1 * time.Second,
		RandomDelay: 1 * time.Second,
	})
	if err != nil {
		return cs, err
	}

	errCh := make(chan error, 1)
	done := make(chan bool)

	c.OnHTML("div[id=card-list3]", processHTML(&cs, errCh, y.source))

	var mu sync.Mutex
	numVisited := 0
	c.OnRequest(func(r *colly.Request) {
		mu.Lock()
		numVisited++
		mu.Unlock()
		fmt.Println("visiting", r.URL.String())
		if numVisited > 100 {
			r.Abort()
			errCh <- errors.New("request limit reached")
		}
	})

	c.OnError(func(_ *colly.Response, err error) {
		errCh <- err
	})

	c.OnScraped(func(_ *colly.Response) {
		done <- true
	})

	go func() {
		err := c.Visit(baseURL.String())
		if err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errCh:
		return nil, err
	case <-done:
		return cs, nil
	}
}

func processHTML(cs *[]*Card, errCh chan error, source string) func(*colly.HTMLElement) {
	return func(e *colly.HTMLElement) {
		rarity := e.ChildText("h3 > span")
		e.ForEach("div[id=card-lits] > div .card-product", func(_ int, el *colly.HTMLElement) {
			var card Card
			card.Code = el.ChildText("span")
			price, err := strconv.ParseInt(extractNumbers(el.ChildText("strong")), 10, 64)
			if err != nil {
				errCh <- err
				return
			}
			card.Price = price
			card.Rarity = rarity
			card.Name = el.ChildText("a > h4")
			card.Source = source
			*cs = append(*cs, &card)
			fmt.Printf("Name: %s, Code: %s Rarity: %s Price: %d\n", card.Name, card.Code, card.Rarity, card.Price)
		})
	}
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
