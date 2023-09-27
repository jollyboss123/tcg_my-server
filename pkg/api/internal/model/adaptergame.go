package model

import "github.com/jollyboss123/tcg_my-server/pkg/game"

func ToGame(game *game.Game) *Game {
	if game == nil {
		return nil
	}

	return &Game{
		Title: game.Title,
		Image: &game.Image,
		Code:  GameCode(game.Code),
	}
}

func ToGames(games []*game.Game) []*Game {
	if len(games) == 0 {
		return make([]*Game, 0)
	}

	var res []*Game
	for _, gg := range games {
		g := ToGame(gg)
		res = append(res, g)
	}
	return res
}
