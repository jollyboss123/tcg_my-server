package detail

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/jollyboss123/tcg_my-server"
	gg "github.com/jollyboss123/tcg_my-server/pkg/game"
	"github.com/jollyboss123/tcg_my-server/pkg/source"
	"log/slog"
	"strings"
)

type opc struct {
	endpoint string
	logger   *slog.Logger
	gs       gg.Service
}

func NewOPC(logger *slog.Logger, gs gg.Service) source.DetailService {
	child := logger.With(slog.String("api", "detail-opc"))
	return &opc{
		endpoint: "https://onepiece-cardgame.dev/cards",
		logger:   child,
		gs:       gs,
	}
}

const (
	searchBar = "[id=searchBar]"
	input     = ".MuiOutlinedInput-root.MuiInputBase-sizeSmall .MuiAutocomplete-input"
	enterKey  = "\r"
	card      = ".MuiBox-root.css-1xdp8rh"
	details   = ".MuiBox-root.css-0"
	firstDiv  = "div:nth-child(1)"
	secDiv    = "div:nth-child(2)"
)

func (o *opc) Fetch(ctx context.Context, code, game string) (*source.DetailInfo, error) {
	_, err := o.gs.Fetch(ctx, game)
	if err != nil {
		o.logger.Error("fetch game", slog.String("error", err.Error()), slog.String("code", code), slog.String("game", game))
		return nil, err
	}
	if !strings.EqualFold(gg.YGO, game) {
		o.logger.Error("check game code", slog.String("error", fmt.Errorf("this detail service does not support %s", game).Error()))
		return nil, err
	}

	code = strings.ToUpper(code)

	options := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36"),
	)

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	var detailNodes []*cdp.Node
	err = chromedp.Run(ctx,
		chromedp.Navigate(o.endpoint),
		chromedp.WaitVisible(searchBar),
		chromedp.SendKeys(input, code),
		chromedp.SendKeys(input, enterKey),
		chromedp.Click(card, chromedp.NodeVisible),
		chromedp.Nodes(details, &detailNodes, chromedp.ByQueryAll),
	)

	if err != nil {
		o.logger.Error("perform automation logic", slog.String("error", err.Error()), slog.String("code", code), slog.String("game", game))
		return nil, err
	}

	var detail source.DetailInfo
	for _, node := range detailNodes {
		err := chromedp.Run(ctx,
			chromedp.ActionFunc(func(ctx context.Context) error {
				titleNode, err := dom.QuerySelector(node.NodeID, firstDiv).Do(ctx)
				if err != nil {
					return err
				}
				valueNode, err := dom.QuerySelector(node.NodeID, secDiv).Do(ctx)
				if err != nil {
					return err
				}
				titleText, err := o.extractText(ctx, titleNode)
				if err != nil {
					return err
				}
				switch titleText {
				case "Name":
					detail.EngName, err = o.extractText(ctx, valueNode)
					return err
				case "Type":
					detail.CardType, err = o.extractText(ctx, valueNode)
					return err
				case "Card Category":
					detail.Category, err = o.extractText(ctx, valueNode)
					return err
				case "Effect":
					detail.Effect, err = o.extractText(ctx, valueNode)
					return err
				case "Product":
					detail.Product, err = o.extractText(ctx, valueNode)
					return err
				case "Color":
					c, err := o.extractText(ctx, valueNode)
					if err != nil {
						return err
					}
					detail.Colors = strings.Split(c, "/")
					return nil
				case "Rarity":
					detail.Rarity, err = o.extractText(ctx, valueNode)
					return err
				case "Life":
					detail.Life, err = o.extractText(ctx, valueNode)
					return err
				case "Power":
					detail.Power, err = o.extractText(ctx, valueNode)
					return err
				case "Attribute":
					detail.Attribute, err = o.extractText(ctx, valueNode)
					return err
				}
				return nil
			}),
		)
		if err != nil {
			o.logger.Error("fetch detail nodes", slog.String("error", err.Error()), slog.String("code", code), slog.String("game", game))
			return nil, err
		}
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
