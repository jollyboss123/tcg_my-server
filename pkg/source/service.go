package source

type Source interface {
	Scrape(code string) (CardInfo, error)
}

type Card struct {
	code      string
	name      string
	rarity    string
	condition string
	price     int64
}

type CardInfo struct {
	url   string
	cards []Card
}
