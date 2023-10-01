package detail

import (
	"context"
	_ "embed"
	"encoding/json"
	er "errors"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/jollyboss123/tcg_my-server"
	"github.com/jollyboss123/tcg_my-server/pkg/api/useragent"
	gg "github.com/jollyboss123/tcg_my-server/pkg/game"
	"github.com/jollyboss123/tcg_my-server/pkg/source"
	"log/slog"
	"regexp"
	"strings"
	"sync"
)

type opc struct {
	logger *slog.Logger
	gs     gg.Service
}

func NewOPC(logger *slog.Logger, gs gg.Service) source.DetailService {
	child := logger.With(slog.String("api", "detail-opc"))
	return &opc{
		logger: child,
		gs:     gs,
	}
}

const (
	SearchBar = "[id=searchBar]"
	Input     = ".MuiOutlinedInput-root.MuiInputBase-sizeSmall .MuiAutocomplete-input"
	EnterKey  = "\r"
	Card      = ".MuiBox-root.css-1xdp8rh"
	Details   = ".MuiBox-root.css-0"
	FirstDiv  = "div:nth-child(1)"
	SecDiv    = "div:nth-child(2)"
)

func (o *opc) Fetch(ctx context.Context, code, game string) (*source.DetailInfo, error) {
	g, err := o.gs.Fetch(ctx, game)
	if err != nil {
		o.logger.Error("fetch game", slog.String("error", err.Error()), slog.String("code", code), slog.String("game", game))
		return nil, err
	}
	if !strings.EqualFold(gg.OPC, game) {
		o.logger.Error("check game code", slog.String("error", fmt.Errorf("this detail service does not support %s", game).Error()))
		return nil, err
	}

	code = strings.ToUpper(code)
	match, err := regexp.MatchString(g.CodeFormat, code)
	if err != nil {
		o.logger.Error("check code format", slog.String("error", err.Error()), slog.String("code", code), slog.String("game", game))
		return nil, err
	}
	if !match {
		o.logger.Error("check code format", slog.String("error", er.New("code format mismatch").Error()), slog.String("code", code), slog.String("game", game))
		return nil, err
	}

	options := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent(useragent.Random()),
	)

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	var detailNodes []*cdp.Node
	err = chromedp.Run(ctx,
		chromedp.Navigate(g.DetailEndpoint),
		chromedp.WaitVisible(SearchBar),
		chromedp.SendKeys(Input, code),
		chromedp.SendKeys(Input, EnterKey),
		chromedp.Click(Card, chromedp.NodeVisible),
		chromedp.Nodes(Details, &detailNodes, chromedp.ByQueryAll),
	)

	if err != nil {
		o.logger.Error("perform automation logic", slog.String("error", err.Error()), slog.String("code", code), slog.String("game", game))
		return nil, err
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	errors := make([]error, 0)

	var detail source.DetailInfo
	for _, node := range detailNodes {
		wg.Add(1)
		go func(node *cdp.Node) {
			defer wg.Done()
			err := chromedp.Run(ctx,
				chromedp.ActionFunc(func(ctx context.Context) error {
					titleNode, err := dom.QuerySelector(node.NodeID, FirstDiv).Do(ctx)
					if err != nil {
						o.logger.Debug("fetch title node", slog.String("error", err.Error()), slog.String("code", code), slog.String("game", game))
						return err
					}
					valueNode, err := dom.QuerySelector(node.NodeID, SecDiv).Do(ctx)
					if err != nil {
						o.logger.Debug("fetch value node", slog.String("error", err.Error()), slog.String("code", code), slog.String("game", game))
						return err
					}
					titleText, err := o.extractText(ctx, titleNode)
					if err != nil {
						o.logger.Debug("fetch title text", slog.String("error", err.Error()), slog.String("code", code), slog.String("game", game))
						return err
					}
					switch titleText {
					case "Name":
						detail.EngName, err = o.extractText(ctx, valueNode)
						o.logger.Debug("fetch name", slog.String("code", code), slog.String("game", game))
						return err
					case "Type":
						detail.CardType, err = o.extractText(ctx, valueNode)
						o.logger.Debug("fetch type", slog.String("code", code), slog.String("game", game))
						return err
					case "Card Category":
						detail.Category, err = o.extractText(ctx, valueNode)
						o.logger.Debug("fetch category", slog.String("code", code), slog.String("game", game))
						return err
					case "Effect":
						detail.Effect, err = o.extractText(ctx, valueNode)
						o.logger.Debug("fetch effect", slog.String("code", code), slog.String("game", game))
						return err
					case "Product":
						detail.Product, err = o.extractText(ctx, valueNode)
						o.logger.Debug("fetch product", slog.String("code", code), slog.String("game", game))
						return err
					case "Color":
						c, err := o.extractText(ctx, valueNode)
						if err != nil {
							o.logger.Debug("fetch color", slog.String("error", err.Error()), slog.String("code", code), slog.String("game", game))
							return err
						}
						detail.Colors = strings.Split(c, "/")
						return nil
					case "Rarity":
						detail.Rarity, err = o.extractText(ctx, valueNode)
						o.logger.Debug("fetch rarity", slog.String("code", code), slog.String("game", game))
						return err
					case "Life":
						detail.Life, err = o.extractText(ctx, valueNode)
						o.logger.Debug("fetch life", slog.String("code", code), slog.String("game", game))
						return err
					case "Power":
						detail.Power, err = o.extractText(ctx, valueNode)
						o.logger.Debug("fetch power", slog.String("code", code), slog.String("game", game))
						return err
					case "Attribute":
						detail.Attribute, err = o.extractText(ctx, valueNode)
						o.logger.Debug("fetch attribute", slog.String("code", code), slog.String("game", game))
						return err
					}
					return nil
				}),
			)
			if err != nil {
				mu.Lock()
				o.logger.Error("fetch detail nodes", slog.String("error", err.Error()), slog.String("code", code), slog.String("game", game))
				errors = append(errors, err)
				mu.Unlock()
			}
		}(node)
	}
	wg.Wait()

	if len(errors) > 0 {
		return nil, errors[0]
	}

	return &detail, nil
}

func (o *opc) extractText(ctx context.Context, nodeID cdp.NodeID) (string, error) {
	r, err := dom.ResolveNode().WithNodeID(nodeID).Do(ctx)
	res, exp, err := runtime.CallFunctionOn(tcg_my.TextJS).WithObjectID(r.ObjectID).Do(ctx)
	if err != nil {
		return "", err
	}
	if exp != nil {
		return "", exp
	}

	var result string
	if res != nil && res.Value != nil {
		if err := json.Unmarshal(res.Value, &result); err != nil {
			return "", err
		}
	}
	return result, nil
}
