package source

import (
	"context"
	"github.com/jollyboss123/tcg_my-server/pkg/currency"
	"github.com/jollyboss123/tcg_my-server/pkg/game"
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
	Game      *game.Game
	Image     string
	Score     int
}

type DetailInfo struct {
	EngName     string
	CardType    string
	Property    string
	Attribute   string
	Types       []string
	Level       string
	LinkArrows  string
	Atk         string
	Def         string
	Link        string
	EffectTypes []string
	Effect      string
	Pendulum    Pendulum
	Status      BanStatus
}

type Pendulum struct {
	EffectTypes []string
	Scale       string
	Effect      string
}

type BanStatus string

const (
	BanStatusUnlimited   BanStatus = "UNLIMITED"
	BanStatusSemiLimited BanStatus = "SEMI_LIMITED"
	BanStatusLimited     BanStatus = "LIMITED"
	BanStatusForbidden   BanStatus = "FORBIDDEN"
)
