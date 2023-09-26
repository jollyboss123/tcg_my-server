// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"
)

type Card struct {
	Code      string      `json:"code"`
	JpName    string      `json:"jpName"`
	Rarity    string      `json:"rarity"`
	Condition *string     `json:"condition,omitempty"`
	Price     int         `json:"price"`
	Source    string      `json:"source"`
	Currency  *Currency   `json:"currency"`
	Image     *string     `json:"image,omitempty"`
	Score     *int        `json:"score,omitempty"`
	Detail    *DetailInfo `json:"detail,omitempty"`
}

type Currency struct {
	Code        string `json:"code"`
	NumericCode string `json:"numericCode"`
	Fraction    int    `json:"fraction"`
	Grapheme    string `json:"grapheme"`
	Template    string `json:"template"`
	Decimal     string `json:"decimal"`
	Thousand    string `json:"thousand"`
}

type DetailInfo struct {
	EngName    *string   `json:"engName,omitempty"`
	CardType   *string   `json:"cardType,omitempty"`
	Property   *string   `json:"property,omitempty"`
	Attribute  *string   `json:"attribute,omitempty"`
	Types      []*string `json:"types,omitempty"`
	Level      *string   `json:"level,omitempty"`
	LinkArrows *string   `json:"linkArrows,omitempty"`
	Attack     *string   `json:"attack,omitempty"`
	Defence    *string   `json:"defence,omitempty"`
	Link       *string   `json:"link,omitempty"`
	Effects    []*string `json:"effects,omitempty"`
	Ability    *string   `json:"ability,omitempty"`
	Pendulum   *Pendulum `json:"pendulum,omitempty"`
	Status     *string   `json:"status,omitempty"`
}

type ExchangeRate struct {
	From *Currency `json:"from"`
	To   *Currency `json:"to"`
	Rate float64   `json:"rate"`
}

type Game struct {
	Title string   `json:"title"`
	Image *string  `json:"image,omitempty"`
	Code  GameCode `json:"code"`
}

type Pendulum struct {
	Effects []*string `json:"effects,omitempty"`
	Scale   *string   `json:"scale,omitempty"`
}

type GameCode string

const (
	GameCodeYgo GameCode = "YGO"
	GameCodePoc GameCode = "POC"
	GameCodeVg  GameCode = "VG"
	GameCodeOpc GameCode = "OPC"
	GameCodeWs  GameCode = "WS"
)

var AllGameCode = []GameCode{
	GameCodeYgo,
	GameCodePoc,
	GameCodeVg,
	GameCodeOpc,
	GameCodeWs,
}

func (e GameCode) IsValid() bool {
	switch e {
	case GameCodeYgo, GameCodePoc, GameCodeVg, GameCodeOpc, GameCodeWs:
		return true
	}
	return false
}

func (e GameCode) String() string {
	return string(e)
}

func (e *GameCode) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = GameCode(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid GameCode", str)
	}
	return nil
}

func (e GameCode) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
