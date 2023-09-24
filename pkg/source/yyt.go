package source

import (
	"context"
	"errors"
	"github.com/gocolly/colly/v2"
	"github.com/jollyboss123/tcg_my-server/pkg/currency"
	"github.com/jollyboss123/tcg_my-server/pkg/game"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

type YYT struct {
	source string
	logger *slog.Logger
	cs     currency.Service
	gs     game.Service
}

func NewYYT(logger *slog.Logger, cs currency.Service, gs game.Service) *YYT {
	child := logger.With(slog.String("api", "yyt"))
	return &YYT{
		source: "Yuyu-tei",
		logger: child,
		cs:     cs,
		gs:     gs,
	}
}

var ErrExceedRequestLimit = errors.New("request limit reached")

func (y *YYT) List(ctx context.Context, query, game string) ([]*Card, error) {
	query = strings.ToUpper(query)
	g, err := y.gs.Fetch(ctx, game)
	if err != nil {
		y.logger.Error("fetch game", slog.String("error", err.Error()), slog.String("query", query), slog.String("game", game))
		return make([]*Card, 0), err
	}
	baseURL, err := url.Parse(g.Endpoint)
	if err != nil {
		y.logger.Error("parsing url", slog.String("error", err.Error()), slog.String("url", g.Endpoint))
		return nil, err
	}

	params := url.Values{}
	params.Add("search_word", query)
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
		y.logger.Error("colly limit rule", slog.String("error", err.Error()))
		return cs, err
	}

	errCh := make(chan error, 1)
	done := make(chan bool)

	c.OnHTML("div[id=card-list3]", y.processHTML(ctx, &cs, errCh, y.source, g.ImageEndpoint, y.logger))

	var mu sync.Mutex
	numVisited := 0
	c.OnRequest(func(r *colly.Request) {
		mu.Lock()
		numVisited++
		mu.Unlock()
		y.logger.Info("scraping start", slog.String("url", r.URL.String()))
		if numVisited > 100 {
			r.Abort()
			y.logger.Error("scraping start", slog.String("error", ErrExceedRequestLimit.Error()), slog.String("url", baseURL.String()))
			errCh <- ErrExceedRequestLimit
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		y.logger.Error("scraping error", slog.String("error", err.Error()), slog.String("url", r.Request.URL.String()))
		errCh <- err
	})

	c.OnScraped(func(r *colly.Response) {
		y.logger.Info("scraping done", slog.String("url", r.Request.URL.String()))
		done <- true
	})

	go func() {
		err := c.Visit(baseURL.String())
		if err != nil {
			y.logger.Error("scraping start", slog.String("error", err.Error()), slog.String("url", baseURL.String()))
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		y.logger.Error("context done", slog.String("error", ctx.Err().Error()), slog.String("url", baseURL.String()))
		return nil, ctx.Err()
	case err := <-errCh:
		return nil, err
	case <-done:
		return cs, nil
	}
}

func (y *YYT) processHTML(ctx context.Context, cs *[]*Card, errCh chan error, source, imageURL string, logger *slog.Logger) func(*colly.HTMLElement) {
	c, err := y.cs.Fetch(ctx, "JPY")
	if err != nil {
		logger.Warn("failed to fetch currency", slog.String("error", err.Error()))
	}
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
			imgURL := el.ChildAttr("a", "href")
			id := strings.Split(imgURL, "/card/")
			if len(id) > 1 {
				card.Image = imageURL + id[1] + ".jpg"
			} else {
				logger.Warn("failed to crawl image", slog.String("error", "no /card/ in url"))
			}
			card.Price = price
			card.Rarity = rarity
			card.Name = el.ChildText("a > h4")
			card.Source = source
			card.Currency = c
			*cs = append(*cs, &card)

			logger.Debug("card info", slog.String("name", card.Name),
				slog.String("code", card.Code),
				slog.String("rarity", card.Rarity),
				slog.String("condition", card.Condition),
				slog.String("img", card.Image),
				slog.Int64("price", card.Price))
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
