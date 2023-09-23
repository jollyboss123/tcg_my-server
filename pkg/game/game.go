package game

import (
	"errors"
	"log/slog"
	"strings"
)

type Games map[string]*Game

var ErrGameNotFound = errors.New("game not found")

func (g Games) GameByCode(code string, logger *slog.Logger) (*Game, error) {
	sc, ok := g[strings.ToUpper(code)]
	if !ok {
		logger.Error("fetch game", slog.String("error", ErrGameNotFound.Error()), slog.String("code", code))
		return nil, ErrGameNotFound
	}

	return sc, nil
}

func (g Games) FetchAll() []*Game {
	var gamesSlice []*Game
	for _, game := range g {
		gamesSlice = append(gamesSlice, game)
	}
	return gamesSlice
}

var games = Games{
	YGO: {Title: "Yu-Gi-Oh!", Image: "https://yuyu-tei.jp/images/gamelogo/ygo.svg"},
	POC: {Title: "Pokemon", Image: "https://yuyu-tei.jp/images/gamelogo/poc.svg"},
	VG:  {Title: "Card Fight!! Vanguard", Image: "https://yuyu-tei.jp/images/gamelogo/vg.svg"},
	OP:  {Title: "One Piece Card Game", Image: "https://yuyu-tei.jp/images/gamelogo/opc.svg"},
	WS:  {Title: "Weiss Schwarz", Image: "https://yuyu-tei.jp/images/gamelogo/ws.svg"},
}
