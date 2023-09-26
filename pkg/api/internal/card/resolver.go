package card

import (
	"context"
	"fmt"
	"github.com/jollyboss123/tcg_my-server/pkg/api/internal/middleware"
	"github.com/jollyboss123/tcg_my-server/pkg/api/internal/model"
	"github.com/jollyboss123/tcg_my-server/pkg/api/internal/resolver"
	"github.com/jollyboss123/tcg_my-server/pkg/source"
)

type cardResolver struct {
	detail source.DetailService
}

func NewCardResolver(d source.DetailService) resolver.CardResolver {
	return &cardResolver{
		detail: d,
	}
}

func (c *cardResolver) Detail(ctx context.Context, obj *model.Card, game model.GameCode) (*model.DetailInfo, error) {
	key := fmt.Sprintf("%s|%s", obj.Code, game.String())
	detailLoader, err := middleware.DetailLoaderFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return detailLoader.Load(ctx, key)()
}
