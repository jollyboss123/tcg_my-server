package model

import "github.com/jollyboss123/tcg_my-server/pkg/currency"

func ToCurrency(currency *currency.Currency) *Currency {
	if currency == nil {
		return nil
	}
	return &Currency{
		Code:        currency.Code,
		NumericCode: currency.NumericCode,
		Fraction:    currency.Fraction,
		Grapheme:    currency.Grapheme,
		Template:    currency.Template,
		Decimal:     currency.Decimal,
		Thousand:    currency.Thousand,
	}
}
