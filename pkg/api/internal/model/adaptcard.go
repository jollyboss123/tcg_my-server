package model

import "github.com/jollyboss123/tcg_my-server/pkg/source"

func ToCard(card *source.Card) *Card {
	if card == nil {
		return nil
	}
	return &Card{
		Code:      card.Code,
		JpName:    card.JpName,
		Rarity:    card.Rarity,
		Condition: &card.Condition,
		Price:     int(card.Price),
		Source:    card.Source,
		Currency:  ToCurrency(card.Currency),
		Image:     &card.Image,
		Score:     &card.Score,
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

func ToDetailInfo(d *source.DetailInfo) *DetailInfo {
	if d == nil {
		return nil
	}
	var types = make([]*string, len(d.Types))
	if d.Types != nil && len(d.Types) > 0 {
		for i := range d.Types {
			types[i] = &d.Types[i]
		}
	}
	var effects = make([]*string, len(d.EffectTypes))
	if d.EffectTypes != nil && len(d.EffectTypes) > 0 {
		for i := range d.EffectTypes {
			effects[i] = &d.EffectTypes[i]
		}
	}
	return &DetailInfo{
		EngName:     &d.EngName,
		CardType:    &d.CardType,
		Property:    &d.Property,
		Attribute:   &d.Attribute,
		Types:       types,
		LinkArrows:  &d.LinkArrows,
		Effect:      &d.Effect,
		Level:       &d.Level,
		Attack:      &d.Atk,
		Defence:     &d.Def,
		Link:        &d.Link,
		EffectTypes: effects,
		Pendulum:    ToPendulum(&d.Pendulum),
		Status:      (*BanStatus)(&d.Status),
	}
}

func ToPendulum(p *source.Pendulum) *Pendulum {
	if p == nil {
		return nil
	}
	var effects = make([]*string, len(p.EffectTypes))
	if p.EffectTypes != nil && len(p.EffectTypes) > 0 {
		for i := range p.EffectTypes {
			effects[i] = &p.EffectTypes[i]
		}
	}

	return &Pendulum{
		EffectTypes: effects,
		Scale:       &p.Scale,
		Effect:      &p.Effect,
	}
}
