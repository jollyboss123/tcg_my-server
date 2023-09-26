package source

import (
	"context"
	"github.com/jollyboss123/tcg_my-server/pkg/currency"
)

type ScrapeService interface {
	List(ctx context.Context, query, game string) ([]*Card, error)
}

type DetailService interface {
	Fetch(ctx context.Context, code, game string) (*DetailInfo, error)
}

type Card struct {
	Code      string
	Name      string
	Rarity    string
	Condition string
	Price     int64
	Source    string
	Currency  *currency.Currency
	Image     string
	Score     int
}

type DetailInfo struct {
	Ability string
}
