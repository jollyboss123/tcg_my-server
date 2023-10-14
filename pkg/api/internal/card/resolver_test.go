package card

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

func TestCardResolver_Detail(t *testing.T) {
	resolver := &internal.Stub{}
	resolver.QueryResolver.Cards = func(ctx context.Context, query string, game model.GameCode) ([]*model.Card, error) {
		return []*model.Card{{}}, nil
	}
	resolver.CardResolver.Detail = func(ctx context.Context, obj *model.Card) (*model.DetailInfo, error) {
		return &model.DetailInfo{
			EngName:   adapt.ToPointer("Centurion Primera"),
			Attribute: adapt.ToPointer("LIGHT"),
			Effect:    adapt.ToPointer("While this card is treated as a Continuous Trap, Level 5 or higher \\\"Centurion\\\" monsters you control cannot be destroyed by card effects. You can only use each of the following effects of \\\"Centurion Primera\\\" once per turn. If this card is Normal or Special Summoned: You can add 1 \\\"Centurion\\\" card from your Deck to your hand, except \\\"Centurion Primera\\\", also you cannot Special Summon \\\"Centurion Primera\\\" for the rest of this turn. During the Main Phase, while this card is treated as a Continuous Trap: You can Special Summon this card.\""),
			EffectTypes: []*string{
				adapt.ToPointer("Continuous-like"),
				adapt.ToPointer("Condition"),
				adapt.ToPointer("Trigger"),
				adapt.ToPointer("Quick-like"),
			},
			CardType: adapt.ToPointer("Monster"),
			Level:    adapt.ToPointer("4"),
			Types: []*string{
				adapt.ToPointer("Spellcaster"),
				adapt.ToPointer("Tuner"),
				adapt.ToPointer("Effect"),
			},
			Status:  adapt.ToPointer(model.BanStatusUnlimited),
			Attack:  adapt.ToPointer("1600"),
			Defence: adapt.ToPointer("1600"),
		}, nil
	}

	c := client.New(handler.NewDefaultServer(
		r.NewExecutableSchema(r.Config{Resolvers: resolver}),
	))

	t.Run("detail query", func(t *testing.T) {
		type response struct {
			Cards []*model.Card
		}
		var resp response
		err := c.Post(`query { cards(query: "dbvs-jp016", game: YGO) { detail { engName attribute effect effectTypes cardType level types status attack defence } } }`, &resp)
		if err != nil {
			t.Fatal(err)
		}

		want := response{
			Cards: []*model.Card{
				{
					Detail: &model.DetailInfo{
						EngName:   adapt.ToPointer("Centurion Primera"),
						Attribute: adapt.ToPointer("LIGHT"),
						Effect:    adapt.ToPointer("While this card is treated as a Continuous Trap, Level 5 or higher \\\"Centurion\\\" monsters you control cannot be destroyed by card effects. You can only use each of the following effects of \\\"Centurion Primera\\\" once per turn. If this card is Normal or Special Summoned: You can add 1 \\\"Centurion\\\" card from your Deck to your hand, except \\\"Centurion Primera\\\", also you cannot Special Summon \\\"Centurion Primera\\\" for the rest of this turn. During the Main Phase, while this card is treated as a Continuous Trap: You can Special Summon this card.\""),
						EffectTypes: []*string{
							adapt.ToPointer("Continuous-like"),
							adapt.ToPointer("Condition"),
							adapt.ToPointer("Trigger"),
							adapt.ToPointer("Quick-like"),
						},
						CardType: adapt.ToPointer("Monster"),
						Level:    adapt.ToPointer("4"),
						Types: []*string{
							adapt.ToPointer("Spellcaster"),
							adapt.ToPointer("Tuner"),
							adapt.ToPointer("Effect"),
						},
						Status:  adapt.ToPointer(model.BanStatusUnlimited),
						Attack:  adapt.ToPointer("1600"),
						Defence: adapt.ToPointer("1600"),
					},
				},
			},
		}

		if !reflect.DeepEqual(want, resp) {
			t.Errorf("got: %v want: %v", resp, want)
		}
	})
}
