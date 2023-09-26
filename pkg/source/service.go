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
	JpName    string
	Rarity    string
	Condition string
	Price     int64
	Source    string
	Currency  *currency.Currency
	Image     string
	Score     int
}

type DetailInfo struct {
	EngName    string
	CardType   string
	Property   string
	Attribute  string
	Types      []string
	Level      string
	LinkArrows string
	Atk        string
	Def        string
	Link       string
	Effects    []string
	Ability    string
	Pendulum   Pendulum
	Status     string
}

type Pendulum struct {
	Effects []string
	Scale   string
}
