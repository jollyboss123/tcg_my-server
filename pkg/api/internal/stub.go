package internal

import (
	"context"
	"github.com/jollyboss123/tcg_my-server/pkg/api/internal/model"
	"github.com/jollyboss123/tcg_my-server/pkg/api/internal/resolver"
)

type Stub struct {
	QueryResolver struct {
		Cards func(ctx context.Context, query string, game model.GameCode) ([]*model.Card, error)
	}
}

type stubCard struct{ *Stub }

func (r *stubCard) Detail(ctx context.Context, obj *model.Card) (*model.DetailInfo, error) {
	return r.Detail(ctx, obj)
}

func (r *Stub) Card() resolver.CardResolver {
	return &stubCard{r}
}

type stubQuery struct{ *Stub }

func (r *stubQuery) Cards(ctx context.Context, query string, game model.GameCode) ([]*model.Card, error) {
	return r.QueryResolver.Cards(ctx, query, game)
}

func (r *stubQuery) Currency(ctx context.Context, code string) (*model.Currency, error) {
	//TODO implement me
	panic("implement me")
}

func (r *stubQuery) ExchangeRate(ctx context.Context, base string, to string) (*model.ExchangeRate, error) {
	//TODO implement me
	panic("implement me")
}

func (r *stubQuery) ExchangeRates(ctx context.Context) ([]*model.ExchangeRate, error) {
	//TODO implement me
	panic("implement me")
}

func (r *stubQuery) Games(ctx context.Context) ([]*model.Game, error) {
	//TODO implement me
	panic("implement me")
}

func (r *Stub) Query() resolver.QueryResolver {
	return &stubQuery{r}
}
