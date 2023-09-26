package detail

import (
	"context"
	"errors"
	"github.com/gocolly/colly/v2"
	"github.com/jollyboss123/tcg_my-server/pkg/source"
	"log/slog"
	"net/url"
	"strings"
	"sync"
	"time"
)

type ygo struct {
	endpoint string
	logger   *slog.Logger
}

func NewYGO(logger *slog.Logger) source.DetailService {
	child := logger.With(slog.String("api", "detail-ygo"))
	return &ygo{
		endpoint: "https://yugipedia.com/wiki/",
		logger:   child,
	}
}

var ErrExceedRequestLimit = errors.New("request limit reached")

func (y ygo) Fetch(ctx context.Context, code, game string) (*source.DetailInfo, error) {
	query := strings.ToUpper(code)
	baseURL, err := url.Parse(y.endpoint + query)
	if err != nil {
		y.logger.Error("parsing url", slog.String("error", err.Error()), slog.String("url", y.endpoint))
		return nil, err
	}

	c := colly.NewCollector(
		colly.AllowedDomains("yugipedia.com"),
		colly.AllowURLRevisit(),
		colly.Async(true),
	)

	err = c.Limit(&colly.LimitRule{
		DomainGlob:  "yugipedia.com/*",
		Delay:       1 * time.Second,
		RandomDelay: 1 * time.Second,
	})
	if err != nil {
		y.logger.Error("colly limit rule", slog.String("error", err.Error()))
		return nil, err
	}

	var lore string
	errCh := make(chan error, 1)
	done := make(chan bool)

	c.OnHTML("div .lore", func(e *colly.HTMLElement) {
		lore = e.ChildText("p")
		y.logger.Info("detail info", slog.String("lore", lore))
	})

	var mu sync.Mutex
	numVisited := 0
	c.OnRequest(func(r *colly.Request) {
		mu.Lock()
		numVisited++
		mu.Unlock()
		y.logger.Info("scraping start", slog.String("url", r.URL.String()))
		if numVisited > 100 {
			r.Abort()
			y.logger.Error("scraping start", slog.String("error", ErrExceedRequestLimit.Error()))
			errCh <- ErrExceedRequestLimit
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		y.logger.Error("scraping error", slog.String("url", r.Request.URL.String()), slog.String("error", err.Error()))
		errCh <- err
	})

	c.OnScraped(func(r *colly.Response) {
		y.logger.Info("scraping done", slog.String("url", r.Request.URL.String()))
		done <- true
	})

	go func() {
		err := c.Visit(baseURL.String())
		if err != nil {
			y.logger.Error("scraping start", slog.String("url", baseURL.String()), slog.String("error", err.Error()))
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		y.logger.Error("context done", slog.String("error", ctx.Err().Error()))
		return nil, ctx.Err()
	case err := <-errCh:
		return nil, err
	case <-done:
		return &source.DetailInfo{Ability: lore}, nil
	}
}
