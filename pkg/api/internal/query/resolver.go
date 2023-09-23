package query

import (
	"context"
	"github.com/jollyboss123/tcg_my-server/pkg/api/internal/model"
	"github.com/jollyboss123/tcg_my-server/pkg/api/internal/resolver"
	"github.com/jollyboss123/tcg_my-server/pkg/currency"
	"github.com/jollyboss123/tcg_my-server/pkg/game"
	"github.com/jollyboss123/tcg_my-server/pkg/rate"
	"github.com/jollyboss123/tcg_my-server/pkg/source"
)

type queryResolver struct {
	scrape   source.ScrapeService
	currency currency.Service
	rate     rate.Service
	game     game.Service
}

func NewQueryResolver(
	s source.ScrapeService,
	c currency.Service,
	r rate.Service,
	g game.Service,
) resolver.QueryResolver {
	return &queryResolver{
		scrape:   s,
		currency: c,
		rate:     r,
		game:     g,
	}
}

func (q queryResolver) Cards(ctx context.Context, query string) ([]*model.Card, error) {
	cards, err := q.scrape.List(ctx, query)
	if err != nil {
		return nil, err
	}
	return model.ToCards(cards), nil
}

func (q queryResolver) Currency(ctx context.Context, code string) (*model.Currency, error) {
	c, err := q.currency.Fetch(ctx, code)
	if err != nil {
		return nil, err
	}
	return model.ToCurrency(c), nil
}

func (q queryResolver) ExchangeRate(ctx context.Context, base, to string) (*model.ExchangeRate, error) {
	r, err := q.rate.Fetch(ctx, base, to)
	if err != nil {
		return nil, err
	}
	return model.ToRate(r), nil
}

func (q queryResolver) ExchangeRates(ctx context.Context) ([]*model.ExchangeRate, error) {
	rs, err := q.rate.List(ctx)
	if err != nil {
		return nil, err
	}
	return model.ToRates(rs), nil
}

func (q queryResolver) Games(ctx context.Context) ([]*model.Game, error) {
	gs, err := q.game.FetchAll(ctx)
	if err != nil {
		return nil, err
	}
	return model.ToGames(gs), nil
}
