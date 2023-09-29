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
	sc.Code = code

	return sc, nil
}

func (g Games) FetchAll() []*Game {
	var gamesSlice []*Game
	for k, game := range g {
		game.Code = k
		gamesSlice = append(gamesSlice, game)
	}
	return gamesSlice
}

var games = Games{
	YGO: {
		Title:         "Yu-Gi-Oh!",
		Image:         "https://yuyu-tei.jp/images/gamelogo/ygo.svg",
		Endpoint:      "https://yuyu-tei.jp/sell/ygo/s/search",
		ImageEndpoint: "https://img.yuyu-tei.jp/card_image/ygo/front/",
	},
	POC: {
		Title:         "Pokemon",
		Image:         "https://yuyu-tei.jp/images/gamelogo/poc.svg",
		Endpoint:      "https://yuyu-tei.jp/sell/poc/s/search",
		ImageEndpoint: "https://img.yuyu-tei.jp/card_image/poc/front/",
	},
	VG: {
		Title:         "Card Fight!! Vanguard",
		Image:         "https://yuyu-tei.jp/images/gamelogo/vg.svg",
		Endpoint:      "https://yuyu-tei.jp/sell/vg/s/search",
		ImageEndpoint: "https://img.yuyu-tei.jp/card_image/vg/front/",
	},
	OPC: {
		Title:         "One Piece Card Game",
		Image:         "https://yuyu-tei.jp/images/gamelogo/opc.svg",
		Endpoint:      "https://yuyu-tei.jp/sell/opc/s/search",
		ImageEndpoint: "https://img.yuyu-tei.jp/card_image/opc/front/",
	},
	WS: {
		Title:          "Weiss Schwarz",
		Image:          "https://yuyu-tei.jp/images/gamelogo/ws.svg",
		Endpoint:       "https://yuyu-tei.jp/sell/ws/s/search",
		ImageEndpoint:  "https://img.yuyu-tei.jp/card_image/ws/front/",
		DetailEndpoint: "https://www.heartofthecards.com/code/cardlist.html?card=WS_",
	},
}
