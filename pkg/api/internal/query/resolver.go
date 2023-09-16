package query

import (
	"context"
	"github.com/jollyboss123/tcg_my-server/pkg/api/internal/model"
	"github.com/jollyboss123/tcg_my-server/pkg/api/internal/resolver"
	"github.com/jollyboss123/tcg_my-server/pkg/source"
)

type queryResolver struct {
	scrape source.ScrapeService
}

func NewQueryResolver(s source.ScrapeService) resolver.QueryResolver {
	return &queryResolver{scrape: s}
}

func (q queryResolver) Cards(ctx context.Context, code string) ([]*model.Card, error) {
	cards, err := q.scrape.List(ctx, code)
	if err != nil {
		return nil, err
	}
	return model.ToCards(cards), nil
}
