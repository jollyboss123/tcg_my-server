package source

import "context"

type ScrapeService interface {
	List(ctx context.Context, code string) ([]*Card, error)
}

type Card struct {
	Code      string
	Name      string
	Rarity    string
	Condition string
	Price     int64
}

type CardInfo struct {
	url   string
	cards []Card
}
