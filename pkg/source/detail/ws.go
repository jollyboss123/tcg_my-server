package detail

import (
	"context"
	"errors"
	"fmt"
	"github.com/gocolly/colly/v2"
	gg "github.com/jollyboss123/tcg_my-server/pkg/game"
	"github.com/jollyboss123/tcg_my-server/pkg/source"
	"log/slog"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"
)

type ws struct {
	logger *slog.Logger
	gs     gg.Service
}

func NewWS(logger *slog.Logger, gs gg.Service) source.DetailService {
	child := logger.With(slog.String("api", "detail-ws"))
	return &ws{
		logger: child,
		gs:     gs,
	}
}

func (w *ws) Fetch(ctx context.Context, code, game string) (*source.DetailInfo, error) {
	g, err := w.gs.Fetch(ctx, game)
	if err != nil {
		w.logger.Error("fetch game", slog.String("error", err.Error()), slog.String("code", code), slog.String("game", game))
		return nil, err
	}
	if !strings.EqualFold(gg.WS, game) {
		w.logger.Error("check game code", slog.String("error", fmt.Errorf("this detail service does not support %s", game).Error()))
		return nil, err
	}

	query := strings.ToUpper(code)
	match, err := regexp.MatchString(g.CodeFormat, query)
	if err != nil {
		w.logger.Error("check code format", slog.String("error", err.Error()), slog.String("code", code), slog.String("game", game))
		return nil, err
	}
	if !match {
		w.logger.Error("check code format", slog.String("error", errors.New("code format mismatch").Error()), slog.String("code", code), slog.String("game", game))
		return nil, err
	}
	baseURL, err := url.Parse(g.DetailEndpoint + query)
	if err != nil {
		w.logger.Error("parsing url", slog.String("error", err.Error()), slog.String("url", g.DetailEndpoint))
		return nil, err
	}

	c := colly.NewCollector(
		colly.AllowedDomains("www.heartofthecards.com"),
		colly.AllowURLRevisit(),
		colly.Async(true),
	)

	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"

	err = c.Limit(&colly.LimitRule{
		DomainGlob:  "www.heartofthecards.com/*",
		Delay:       1 * time.Second,
		RandomDelay: 1 * time.Second,
	})
	if err != nil {
		w.logger.Error("colly limit rule", slog.String("error", err.Error()))
		return nil, err
	}

	var detail source.DetailInfo
	errCh := make(chan error, 1)
	done := make(chan bool)
	isDetailAvailable := false

	c.OnHTML("div.tcgrcontainer", func(e *colly.HTMLElement) {
		table := e.DOM.NextAllFiltered("table").First()
		detail.EngName = table.Find("b").First().Text()
	})

	c.OnHTML("table .cards", func(e *colly.HTMLElement) {
		isDetailAvailable = true
		e.ForEach("tr", func(_ int, el *colly.HTMLElement) {
			titleNodes := el.ChildTexts(".cards")
			valueNodes := el.ChildTexts(".cards2")

			type KeywordAction struct {
				Keyword string
				Action  func(string)
			}

			actions := []KeywordAction{
				{"Rarity", func(v string) { detail.Rarity = v }},
				{"Type", func(v string) { detail.CardType = v }},
				{"Level", func(v string) { detail.Level = v }},
				{"Power", func(v string) { detail.Power = v }},
				{"Cost", func(v string) { detail.Cost = v }},
				{"Triggers", func(v string) {
					trigger := extractNumbers(v)
					if trigger == "" {
						trigger = "0"
					}
					detail.Trigger = trigger
				},
				},
				{"Soul", func(v string) { detail.Soul = v }},
				{"Trait", func(v string) {
					if v != "None" {
						detail.Traits = append(detail.Traits, v)
					}
				},
				},
			}

			for i, value := range valueNodes {
				for _, action := range actions {
					if strings.Contains(titleNodes[i], action.Keyword) {
						action.Action(value)
					}
				}
			}

			if strings.Contains(el.ChildText(".cards"), "English Card Text") {
				detail.Effect = el.DOM.Next().Text()
			}
		})
	})

	var mu sync.Mutex
	numVisited := 0
	c.OnRequest(func(r *colly.Request) {
		mu.Lock()
		numVisited++
		mu.Unlock()
		w.logger.Info("scraping start", slog.String("url", r.URL.String()))
		if numVisited > 100 {
			r.Abort()
			w.logger.Error("scraping start", slog.String("error", ErrExceedRequestLimit.Error()))
			errCh <- ErrExceedRequestLimit
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		w.logger.Error("scraping error", slog.String("url", r.Request.URL.String()), slog.String("error", err.Error()))
		errCh <- err
	})

	c.OnScraped(func(r *colly.Response) {
		w.logger.Info("scraping done", slog.String("url", r.Request.URL.String()))
		if !isDetailAvailable {
			w.logger.Error("detail available", slog.String("error", fmt.Errorf("no details found for: %s", code).Error()), slog.String("code", code), slog.String("game", game))
			errCh <- fmt.Errorf("no details found for: %s", code)
		}
		done <- true
	})

	go func() {
		err := c.Visit(baseURL.String())
		if err != nil {
			w.logger.Error("scraping start", slog.String("url", baseURL.String()), slog.String("error", err.Error()))
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		w.logger.Error("context done", slog.String("error", ctx.Err().Error()))
		return nil, ctx.Err()
	case err := <-errCh:
		return nil, err
	case <-done:
		return &detail, nil
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
