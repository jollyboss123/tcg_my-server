package internal

import (
	"context"
	"github.com/jollyboss123/tcg_my-server/pkg/api/internal/model"
	"github.com/jollyboss123/tcg_my-server/pkg/api/internal/resolver"
)

type Stub struct {
	CardResolver struct {
		Detail func(ctx context.Context, obj *model.Card) (*model.DetailInfo, error)
	}
	QueryResolver struct {
		Cards         func(ctx context.Context, query string, game model.GameCode) ([]*model.Card, error)
		Currency      func(ctx context.Context, code string) (*model.Currency, error)
		ExchangeRate  func(ctx context.Context, base, to string) (*model.ExchangeRate, error)
		ExchangeRates func(ctx context.Context) ([]*model.ExchangeRate, error)
		Games         func(ctx context.Context) ([]*model.Game, error)
	}
}

type stubCard struct{ *Stub }

func (r *stubCard) Detail(ctx context.Context, obj *model.Card) (*model.DetailInfo, error) {
	return r.CardResolver.Detail(ctx, obj)
}

func (r *Stub) Card() resolver.CardResolver {
	return &stubCard{r}
}

type stubQuery struct{ *Stub }

func (r *stubQuery) Cards(ctx context.Context, query string, game model.GameCode) ([]*model.Card, error) {
	return r.QueryResolver.Cards(ctx, query, game)
}

func (r *stubQuery) Currency(ctx context.Context, code string) (*model.Currency, error) {
	return r.QueryResolver.Currency(ctx, code)
}

func (r *stubQuery) ExchangeRate(ctx context.Context, base string, to string) (*model.ExchangeRate, error) {
	return r.QueryResolver.ExchangeRate(ctx, base, to)
}

func (r *stubQuery) ExchangeRates(ctx context.Context) ([]*model.ExchangeRate, error) {
	return r.QueryResolver.ExchangeRates(ctx)
}

func (r *stubQuery) Games(ctx context.Context) ([]*model.Game, error) {
	return r.QueryResolver.Games(ctx)
}

func (r *Stub) Query() resolver.QueryResolver {
	return &stubQuery{r}
}
