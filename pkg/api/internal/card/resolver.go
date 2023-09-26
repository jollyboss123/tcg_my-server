package card

import (
	"context"
	"github.com/jollyboss123/tcg_my-server/pkg/api/internal/model"
	"github.com/jollyboss123/tcg_my-server/pkg/api/internal/resolver"
	"github.com/jollyboss123/tcg_my-server/pkg/source"
)

type cardResolver struct {
	//scrape source.ScrapeService
	detail source.DetailService
}

func NewCardResolver(d source.DetailService) resolver.CardResolver {
	return &cardResolver{
		//scrape: s,
		detail: d,
	}
}

func (c *cardResolver) Detail(ctx context.Context, obj *model.Card, game model.GameCode) (*model.DetailInfo, error) {
	d, err := c.detail.Fetch(ctx, obj.Code, game.String())
	if err != nil {
		return nil, err
	}
	return model.ToDetailInfo(d), nil
}
