package query

import (
	"context"
	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/jollyboss123/tcg_my-server/pkg/api/internal"
	"github.com/jollyboss123/tcg_my-server/pkg/api/internal/model"
	r "github.com/jollyboss123/tcg_my-server/pkg/api/internal/resolver"
	"reflect"
	"testing"
)

func TestCardsQuery(t *testing.T) {
	resolver := &internal.Stub{}
	resolver.QueryResolver.Cards = func(ctx context.Context, query string, game model.GameCode) ([]*model.Card, error) {
		return []*model.Card{
			{
				Code:   "DBVS-JP016",
				JpName: "重騎士プリメラ",
				Rarity: "QCSE",
			},
			{
				Code:   "DBVS-JP016",
				JpName: "重騎士プリメラ",
				Rarity: "SR",
			},
		}, nil
	}

	c := client.New(handler.NewDefaultServer(
		r.NewExecutableSchema(r.Config{Resolvers: resolver}),
	))

	type card struct {
		Code   string
		JpName string
		Rarity string
	}

	t.Run("cards query", func(t *testing.T) {
		type response struct {
			Cards []card
		}

		var resp response
		err := c.Post(`query { cards(query: "DBVS-JP016", game: YGO) { code jpName rarity } }`, &resp)
		if err != nil {
			t.Fatal(err)
		}

		want := response{
			Cards: []card{
				{
					Code:   "DBVS-JP016",
					JpName: "重騎士プリメラ",
					Rarity: "QCSE",
				},
				{
					Code:   "DBVS-JP016",
					JpName: "重騎士プリメラ",
					Rarity: "SR",
				},
			},
		}

		if !reflect.DeepEqual(want, resp) {
			t.Errorf("got: %v want: %v", resp, want)
		}
	})

	t.Run("invalid game code query", func(t *testing.T) {
		type response struct {
			Cards []card
		}

		var resp response
		err := c.Post(`query { cards(query: "DBVS-JP016", game: ABC) { code jpName rarity } }`, &resp)
		if err == nil {
			t.Error("expected error here")
		}
	})
}
