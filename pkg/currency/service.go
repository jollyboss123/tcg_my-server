package currency

import (
	"context"
	"log/slog"
)

type Service interface {
	Fetch(ctx context.Context, code string) (*Currency, error)
}

type Currency struct {
	Code        string
	NumericCode string
	Fraction    int
	Grapheme    string
	Template    string
	Decimal     string
	Thousand    string
}

type service struct {
	logger *slog.Logger
}

func NewService(logger *slog.Logger) Service {
	child := logger.With(slog.String("api", "currency"))
	return &service{
		logger: child,
	}
}

func (s *service) Fetch(ctx context.Context, code string) (*Currency, error) {
	return currencies.CurrencyByCode(code, s.logger)
}
