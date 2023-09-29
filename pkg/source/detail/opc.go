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
	"github.com/jollyboss123/tcg_my-server/pkg/game"
	"github.com/jollyboss123/tcg_my-server/pkg/source"
	"log"
	"log/slog"
)

type opc struct {
	endpoint string
	logger   *slog.Logger
	gs       game.Service
}

func NewOPC(logger *slog.Logger, gs game.Service) source.DetailService {
	child := logger.With(slog.String("api", "detail-opc"))
	return &opc{
		endpoint: "https://onepiece-cardgame.dev/cards",
		logger:   child,
		gs:       gs,
	}
}

const enterKey = "\r"

func (o *opc) Fetch(ctx context.Context, code, game string) (*source.DetailInfo, error) {
	ctx, cancel := chromedp.NewContext(
		context.Background(),
	)
	defer cancel()

	var detailNodes []*cdp.Node
	err := chromedp.Run(ctx,
		chromedp.Navigate(o.endpoint),
		chromedp.WaitVisible("[id=searchBar]"),
		chromedp.SendKeys(".MuiOutlinedInput-root.MuiInputBase-sizeSmall .MuiAutocomplete-input", code),
		chromedp.SendKeys(".MuiOutlinedInput-root.MuiInputBase-sizeSmall .MuiAutocomplete-input", enterKey),
		chromedp.Click(".MuiBox-root.css-1xdp8rh", chromedp.NodeVisible),
		chromedp.Nodes(".MuiBox-root.css-0", &detailNodes, chromedp.ByQueryAll),
	)

	if err != nil {
		log.Fatal("Error while performing the automation logic:", err)
	}

	for _, node := range detailNodes {
		err := chromedp.Run(ctx,
			chromedp.ActionFunc(func(ctx context.Context) error {
				// Get the text content from the first div
				titleNode, err := dom.QuerySelector(node.NodeID, "div:nth-child(1)").Do(ctx)
				if err != nil {
					return err
				}
				titleText, err := o.extractText(ctx, titleNode)
				if err != nil {
					return err
				}

				fmt.Println(titleText)

				//Get the text content from the second div
				valueNode, err := dom.QuerySelector(node.NodeID, "div:nth-child(2)").Do(ctx)
				valueText, err := o.extractText(ctx, valueNode)

				fmt.Println(valueText)
				return nil
			}),
		)
		if err != nil {
			// Handle error
			fmt.Println(err)
		}
	}

	if err != nil {
		o.logger.Error("scrape", slog.String("error", err.Error()))
	}

	return nil, nil
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
