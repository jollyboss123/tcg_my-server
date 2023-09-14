package source

type Source interface {
	Scrape(code string) (CardInfo, error)
}

type Card struct {
	id        string
	jpName    string
	rarity    string
	condition string
	price     int64
}

type CardInfo struct {
	url   string
	cards []Card
}
