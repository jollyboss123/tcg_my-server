package source

import "context"

type ScrapeService interface {
	List(ctx context.Context, query string) ([]*Card, error)
}

type Card struct {
	Code      string
	Name      string
	Rarity    string
	Condition string
	Price     int64
	Source    string
	Currency  string
}
