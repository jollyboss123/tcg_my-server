package model

import "github.com/jollyboss123/tcg_my-server/pkg/rate"

func ToRate(rate *rate.ExchangeRate) *ExchangeRate {
	if rate == nil {
		return nil
	}
	return &ExchangeRate{
		From: ToCurrency(rate.From),
		To:   ToCurrency(rate.To),
		Rate: rate.Rate,
	}
}

func ToRates(rates []*rate.ExchangeRate) []*ExchangeRate {
	if len(rates) == 0 {
		return make([]*ExchangeRate, 0)
	}

	var res []*ExchangeRate
	for _, rate := range rates {
		r := ToRate(rate)
		res = append(res, r)
	}
	return res
}
