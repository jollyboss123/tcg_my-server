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

	var detail source.DetailInfo
	errCh := make(chan error, 1)
	done := make(chan bool)

	c.OnHTML("div .heading", func(e *colly.HTMLElement) {
		detail.EngName = e.Text
	})

	c.OnHTML("div .card-table-columns", func(e *colly.HTMLElement) {
		isPendulum := false
		e.ForEach("tr", func(_ int, el *colly.HTMLElement) {
			header := el.ChildText("th")
			switch header {
			case "Card type":
				detail.CardType = el.ChildText("td > a")
			case "Attribute":
				detail.Attribute = el.ChildText("td > a")
			case "Property":
				detail.Property = el.ChildText("td > a")
			case "Types":
				el.ForEach("td > a", func(_ int, ele *colly.HTMLElement) {
					detail.Types = append(detail.Types, ele.Text)
				})
			case "Link Arrows":
				el.ForEach("td > div > div:nth-child(2)", func(_ int, ele *colly.HTMLElement) {
					detail.LinkArrows = ele.ChildText("a")
				})
			case "Pendulum Scale":
				isPendulum = true
				detail.Pendulum.Scale = el.ChildText("td > a:nth-child(2)")
			case "ATK / LINK":
				p := el.ChildTexts("td > a")
				detail.Atk = p[0]
				detail.Link = p[1]
			case "ATK / DEF":
				p := el.ChildTexts("td > a")
				detail.Atk = p[0]
				detail.Def = p[1]
			case "Level":
				fallthrough
			case "Rank":
				detail.Level = el.ChildText("td > a:nth-child(1)")
			case "Effect types":
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
			case "Status":
				el.ForEach("i", func(_ int, ele *colly.HTMLElement) {
					if strings.TrimSpace(ele.Text) == "OCG" {
						detail.Status = source.BanStatus(ele.DOM.Prev().Text())
					}
				})
			}
		})

		if !isPendulum {
			detail.Effect = e.ChildText("div .lore p")
		} else {
			e.ForEach("div .lore dd", func(_ int, el *colly.HTMLElement) {
				dtText := el.DOM.Prev().Text()

				isPendulumEffect := strings.Contains(dtText, "Pendulum")

				if isPendulumEffect {
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
