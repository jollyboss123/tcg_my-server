package game

import (
	"context"
	"log/slog"
)

type Service interface {
	Fetch(ctx context.Context, code string) (*Game, error)
	FetchAll(ctx context.Context) ([]*Game, error)
}

type Game struct {
	Title         string
	Image         string
	Code          string
	Endpoint      string
	ImageEndpoint string
}

type service struct {
	logger *slog.Logger
}

func NewService(logger *slog.Logger) Service {
	child := logger.With(slog.String("api", "game"))
	return &service{
		logger: child,
	}
}

func (s *service) Fetch(ctx context.Context, code string) (*Game, error) {
	return games.GameByCode(code, s.logger)
}

func (s *service) FetchAll(ctx context.Context) ([]*Game, error) {
	return games.FetchAll(), nil
}
