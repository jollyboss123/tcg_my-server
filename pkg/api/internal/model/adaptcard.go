package model

import "github.com/jollyboss123/tcg_my-server/pkg/source"

func ToCard(card *source.Card) *Card {
	if card == nil {
		return nil
	}
	return &Card{
		Code:      card.Code,
		Name:      card.Name,
		Rarity:    card.Rarity,
		Condition: &card.Condition,
		Price:     int(card.Price),
		Source:    card.Source,
		Currency:  ToCurrency(card.Currency),
		Image:     &card.Image,
	}
}

func ToCards(cards []*source.Card) []*Card {
	if len(cards) == 0 {
		return make([]*Card, 0)
	}

	var res []*Card
	for _, card := range cards {
		c := ToCard(card)
		res = append(res, c)
	}
	return res
}
