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
	"strings"
	"sync"
	"time"
)

type ygo struct {
	endpoint string
	logger   *slog.Logger
	gs       gg.Service
}

func NewYGO(logger *slog.Logger, gs gg.Service) source.DetailService {
	child := logger.With(slog.String("api", "detail-ygo"))
	return &ygo{
		endpoint: "https://yugipedia.com/wiki/",
		logger:   child,
		gs:       gs,
	}
}

var ErrExceedRequestLimit = errors.New("request limit reached")

func (y ygo) Fetch(ctx context.Context, code, game string) (*source.DetailInfo, error) {
	_, err := y.gs.Fetch(ctx, game)
	if err != nil {
		y.logger.Error("fetch game", slog.String("error", err.Error()), slog.String("code", code), slog.String("game", game))
		return nil, err
	}
	if !strings.EqualFold(gg.YGO, game) {
		y.logger.Error("check game code", slog.String("error", fmt.Errorf("this detail service does not support %s", game).Error()))
		return nil, err
	}

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

	// setting a valid User-Agent header
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"

	err = c.Limit(&colly.LimitRule{
		DomainGlob:  "yugipedia.com/*",
		Delay:       1 * time.Second,
		RandomDelay: 1 * time.Second,
	})
	if err != nil {
		y.logger.Error("colly limit rule", slog.String("error", err.Error()))
		return nil, err
	}

	var detail source.DetailInfo
	errCh := make(chan error, 1)
	done := make(chan bool)

	c.OnHTML("div .heading", func(e *colly.HTMLElement) {
		detail.EngName = e.Text
	})

	c.OnHTML("div .card-table-columns", func(e *colly.HTMLElement) {
		var isPendulum bool

		fieldResolvers := map[string]func(el *colly.HTMLElement){
			"Card type": func(el *colly.HTMLElement) { detail.CardType = el.ChildText("td > a") },
			"Attribute": func(el *colly.HTMLElement) { detail.Attribute = el.ChildText("td > a") },
			"Property": func(el *colly.HTMLElement) {
				detail.Property = el.ChildText("td > a")
			},
			"Types": func(el *colly.HTMLElement) {
				el.ForEach("td > a", func(_ int, ele *colly.HTMLElement) {
					detail.Types = append(detail.Types, ele.Text)
				})
			},
			"Link Arrows": func(el *colly.HTMLElement) {
				el.ForEach("td > div > div:nth-child(2)", func(_ int, ele *colly.HTMLElement) {
					detail.LinkArrows = ele.ChildText("a")
				})
			},
			"Pendulum Scale": func(el *colly.HTMLElement) {
				isPendulum = true
				detail.Pendulum.Scale = el.ChildText("td > a:nth-child(2)")
			},
			"ATK / LINK": func(el *colly.HTMLElement) {
				p := el.ChildTexts("td > a")
				detail.Atk, detail.Link = p[0], p[1]
			},
			"ATK / DEF": func(el *colly.HTMLElement) {
				p := el.ChildTexts("td > a")
				detail.Atk, detail.Def = p[0], p[1]
			},
			"Level": func(el *colly.HTMLElement) {
				detail.Level = el.ChildText("td > a:nth-child(1)")
			},
			"Rank": func(el *colly.HTMLElement) {
				detail.Level = el.ChildText("td > a:nth-child(1)")
			},
			"Effect types": func(el *colly.HTMLElement) {
				if !isPendulum {
					el.ForEach("li", func(_ int, ele *colly.HTMLElement) {
						detail.EffectTypes = append(detail.EffectTypes, ele.ChildText("a"))
					})
				} else {
					el.ForEach("ul", func(_ int, ele *colly.HTMLElement) {
						dtText := ele.DOM.Prev().Text()

						isPendulumEffect := strings.Contains(dtText, "Pendulum")

						ele.ForEach("li > a", func(_ int, elem *colly.HTMLElement) {
							if isPendulumEffect {
								detail.Pendulum.EffectTypes = append(detail.Pendulum.EffectTypes, elem.Text)
							} else {
								detail.EffectTypes = append(detail.EffectTypes, elem.Text)
							}
						})
					})
				}
			},
			"Status": func(el *colly.HTMLElement) {
				el.ForEach("i", func(_ int, ele *colly.HTMLElement) {
					if strings.TrimSpace(ele.Text) == "OCG" {
						detail.Status = source.BanStatus(ele.DOM.Prev().Text())
					}
				})
			},
		}

		e.ForEach("tr", func(_ int, el *colly.HTMLElement) {
			header := el.ChildText("th")
			if resolver, exists := fieldResolvers[header]; exists {
				resolver(el)
			}
		})

		detail.Effect = e.ChildText("div .lore p")

		if isPendulum {
			e.ForEach("div .lore dd", func(_ int, el *colly.HTMLElement) {
				if strings.Contains(el.DOM.Prev().Text(), "Pendulum") {
					detail.Pendulum.Effect = el.Text
				} else {
					detail.Effect = el.Text
				}
			})
		}
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
		return &detail, nil
	}
}
