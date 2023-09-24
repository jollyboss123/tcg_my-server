package model

import "github.com/jollyboss123/tcg_my-server/pkg/game"

func ToGame(game *game.Game) *Game {
	if game == nil {
		return nil
	}

	return &Game{
		Title: game.Title,
		Image: &game.Image,
	}
}

func ToGames(games []*game.Game) []*Game {
	if len(games) == 0 {
		return make([]*Game, 0)
	}

	var res []*Game
	for _, game := range games {
		g := ToGame(game)
		res = append(res, g)
	}
	return res
}