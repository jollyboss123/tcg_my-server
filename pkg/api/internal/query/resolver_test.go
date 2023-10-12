package query

import (
	"context"
	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/jollyboss123/tcg_my-server/pkg/api/internal"
	"github.com/jollyboss123/tcg_my-server/pkg/api/internal/model"
	r "github.com/jollyboss123/tcg_my-server/pkg/api/internal/resolver"
	"github.com/jollyboss123/tcg_my-server/pkg/internal/adapt"
	"reflect"
	"testing"
)

func TestQueryResolver_Cards(t *testing.T) {
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

	t.Run("cards query", func(t *testing.T) {
		type response struct {
			Cards []*model.Card
		}

		var resp response
		err := c.Post(`query { cards(query: "DBVS-JP016", game: YGO) { code jpName rarity } }`, &resp)
		if err != nil {
			t.Fatal(err)
		}

		want := response{
			Cards: []*model.Card{
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
			Cards []*model.Card
		}

		var resp response
		err := c.Post(`query { cards(query: "DBVS-JP016", game: ABC) { code jpName rarity } }`, &resp)
		if err == nil {
			t.Error("expected error here")
		}
	})
}

func TestQueryResolver_Currency(t *testing.T) {
	resolver := &internal.Stub{}
	resolver.QueryResolver.Currency = func(ctx context.Context, code string) (*model.Currency, error) {
		return &model.Currency{
			Code:        "MYR",
			NumericCode: "458",
			Fraction:    2,
			Grapheme:    "RM",
			Template:    "$1",
			Decimal:     ".",
			Thousand:    ",",
		}, nil
	}

	c := client.New(handler.NewDefaultServer(
		r.NewExecutableSchema(r.Config{Resolvers: resolver}),
	))

	t.Run("currency query", func(t *testing.T) {
		type response struct {
			Currency model.Currency
		}

		var resp response
		err := c.Post(`query { currency(code: "MYR") { code numericCode fraction grapheme template decimal thousand } }`, &resp)
		if err != nil {
			t.Fatal(err)
		}

		want := response{
			Currency: model.Currency{
				Code:        "MYR",
				NumericCode: "458",
				Fraction:    2,
				Grapheme:    "RM",
				Template:    "$1",
				Decimal:     ".",
				Thousand:    ",",
			},
		}

		if !reflect.DeepEqual(want, resp) {
			t.Errorf("got: %v want: %v", resp, want)
		}
	})
}

func TestQueryResolver_ExchangeRate(t *testing.T) {
	resolver := &internal.Stub{}
	resolver.QueryResolver.ExchangeRate = func(ctx context.Context, base, to string) (*model.ExchangeRate, error) {
		return &model.ExchangeRate{
			From: &model.Currency{
				Code:        "MYR",
				NumericCode: "458",
				Fraction:    2,
				Grapheme:    "RM",
				Template:    "$1",
				Decimal:     ".",
				Thousand:    ",",
			},
			To: &model.Currency{
				Code:        "MYR",
				NumericCode: "458",
				Fraction:    2,
				Grapheme:    "RM",
				Template:    "$1",
				Decimal:     ".",
				Thousand:    ",",
			},
			Rate: 1.0,
		}, nil
	}

	c := client.New(handler.NewDefaultServer(
		r.NewExecutableSchema(r.Config{Resolvers: resolver}),
	))

	t.Run("exchange rate query", func(t *testing.T) {
		type response struct {
			ExchangeRate model.ExchangeRate
		}

		var resp response
		err := c.Post(`query { exchangeRate(base: "MYR", to: "MYR") { from { code numericCode fraction grapheme template decimal thousand } to { code numericCode fraction grapheme template decimal thousand } rate } }`, &resp)
		if err != nil {
			t.Fatal(err)
		}

		want := response{
			ExchangeRate: model.ExchangeRate{
				From: &model.Currency{
					Code:        "MYR",
					NumericCode: "458",
					Fraction:    2,
					Grapheme:    "RM",
					Template:    "$1",
					Decimal:     ".",
					Thousand:    ",",
				}, To: &model.Currency{
					Code:        "MYR",
					NumericCode: "458",
					Fraction:    2,
					Grapheme:    "RM",
					Template:    "$1",
					Decimal:     ".",
					Thousand:    ",",
				}, Rate: 1.0},
		}

		if !reflect.DeepEqual(want, resp) {
			t.Errorf("got: %v want: %v", resp, want)
		}
	})
}

func TestQueryResolver_ExchangeRates(t *testing.T) {
	resolver := &internal.Stub{}
	resolver.QueryResolver.ExchangeRates = func(ctx context.Context) ([]*model.ExchangeRate, error) {
		return []*model.ExchangeRate{
			{
				From: &model.Currency{
					Code:        "EUR",
					NumericCode: "978",
					Fraction:    2,
					Grapheme:    "\u20ac",
					Template:    "$1",
					Decimal:     ".",
					Thousand:    ",",
				},
				To: &model.Currency{
					Code:        "MYR",
					NumericCode: "458",
					Fraction:    2,
					Grapheme:    "RM",
					Template:    "$1",
					Decimal:     ".",
					Thousand:    ",",
				},
				Rate: 5.012907,
			},
			{
				From: &model.Currency{
					Code:        "EUR",
					NumericCode: "978",
					Fraction:    2,
					Grapheme:    "\u20ac",
					Template:    "$1",
					Decimal:     ".",
					Thousand:    ",",
				},
				To: &model.Currency{
					Code:        "USD",
					NumericCode: "840",
					Fraction:    2,
					Grapheme:    "$",
					Template:    "$1",
					Decimal:     ".",
					Thousand:    ",",
				},
				Rate: 1.062395,
			},
		}, nil
	}

	c := client.New(handler.NewDefaultServer(
		r.NewExecutableSchema(r.Config{Resolvers: resolver}),
	))

	t.Run("exchange rates query", func(t *testing.T) {
		type response struct {
			ExchangeRates []*model.ExchangeRate
		}
		var resp response

		err := c.Post(`query { exchangeRates { from { code numericCode fraction grapheme template decimal thousand } to { code numericCode fraction grapheme template decimal thousand } rate } }`, &resp)
		if err != nil {
			t.Fatal(err)
		}

		want := response{
			ExchangeRates: []*model.ExchangeRate{
				{
					From: &model.Currency{
						Code:        "EUR",
						NumericCode: "978",
						Fraction:    2,
						Grapheme:    "\u20ac",
						Template:    "$1",
						Decimal:     ".",
						Thousand:    ",",
					},
					To: &model.Currency{
						Code:        "MYR",
						NumericCode: "458",
						Fraction:    2,
						Grapheme:    "RM",
						Template:    "$1",
						Decimal:     ".",
						Thousand:    ",",
					},
					Rate: 5.012907,
				},
				{
					From: &model.Currency{
						Code:        "EUR",
						NumericCode: "978",
						Fraction:    2,
						Grapheme:    "\u20ac",
						Template:    "$1",
						Decimal:     ".",
						Thousand:    ",",
					},
					To: &model.Currency{
						Code:        "USD",
						NumericCode: "840",
						Fraction:    2,
						Grapheme:    "$",
						Template:    "$1",
						Decimal:     ".",
						Thousand:    ",",
					},
					Rate: 1.062395,
				},
			},
		}

		if !reflect.DeepEqual(want, resp) {
			t.Errorf("got: %v want: %v", resp, want)
		}
	})
}

func TestQueryResolver_Games(t *testing.T) {
	resolver := &internal.Stub{}
	resolver.QueryResolver.Games = func(ctx context.Context) ([]*model.Game, error) {
		return []*model.Game{
			{
				Title: "Yu-Gi-Oh!",
				Image: adapt.ToPointer("https://yuyu-tei.jp/images/gamelogo/ygo.svg"),
				Code:  "YGO",
			},
			{
				Title: "One Piece Card Game",
				Image: adapt.ToPointer("https://yuyu-tei.jp/images/gamelogo/opc.svg"),
				Code:  "OPC",
			},
			{
				Title: "Weiss Schwarz",
				Image: adapt.ToPointer("https://yuyu-tei.jp/images/gamelogo/ws.svg"),
				Code:  "WS",
			},
		}, nil
	}

	c := client.New(handler.NewDefaultServer(
		r.NewExecutableSchema(r.Config{Resolvers: resolver}),
	))

	t.Run("games query", func(t *testing.T) {
		type response struct {
			Games []*model.Game
		}
		var resp response
		err := c.Post(`query { games { title image code } }`, &resp)
		if err != nil {
			t.Fatal(err)
		}

		want := response{
			Games: []*model.Game{
				{
					Title: "Yu-Gi-Oh!",
					Image: adapt.ToPointer("https://yuyu-tei.jp/images/gamelogo/ygo.svg"),
					Code:  "YGO",
				},
				{
					Title: "One Piece Card Game",
					Image: adapt.ToPointer("https://yuyu-tei.jp/images/gamelogo/opc.svg"),
					Code:  "OPC",
				},
				{
					Title: "Weiss Schwarz",
					Image: adapt.ToPointer("https://yuyu-tei.jp/images/gamelogo/ws.svg"),
					Code:  "WS",
				},
			},
		}

		if !reflect.DeepEqual(want, resp) {
			t.Errorf("got: %v want: %v", resp, want)
		}
	})
}
