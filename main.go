package main

import "tcg_my/pricefinder"

func main() {
	finder := pricefinder.NewBigWebPriceFinder()
	finder.FindPrices("ユニオン・キャリアー")
	//fmt.Println("starting")
	//// initializing the slice of structs to store the data to scrape
	//var pokemonProducts []PokemonProduct
	//
	//// creating a new Colly instance
	//c := colly.NewCollector()
	//
	//// scraping logic
	//c.OnHTML("li.product", func(e *colly.HTMLElement) {
	//	pokemonProduct := PokemonProduct{}
	//
	//	pokemonProduct.url = e.ChildAttr("a", "href")
	//	pokemonProduct.image = e.ChildAttr("img", "src")
	//	pokemonProduct.name = e.ChildText("h2")
	//	pokemonProduct.price = e.ChildText(".price")
	//
	//	pokemonProducts = append(pokemonProducts, pokemonProduct)
	//	fmt.Println(pokemonProducts)
	//})
	//
	//// visiting the target page
	//err := c.Visit("https://scrapeme.live/shop/")
	//if err != nil {
	//	log.Printf("failed to visit url: %v\n", err)
	//	return
	//}

	//c.OnRequest(func(r *colly.Request) {
	//	fmt.Println("Visiting: ", r.URL)
	//})
	//c.OnError(func(_ *colly.Response, err error) {
	//	log.Println("something went wrong: ", err)
	//})
	//c.OnResponse(func(r *colly.Response) {
	//	fmt.Println("Page visited: ", r.Request.URL)
	//})
	//c.OnHTML("a", func(e *colly.HTMLElement) {
	//	fmt.Println("%v", e.Attr("href"))
	//})
	//c.OnScraped(func(r *colly.Response) {
	//	fmt.Println(r.Request.URL, " scraped!")
	//})

	select {}
}

type PokemonProduct struct {
	url, image, name, price string
}
