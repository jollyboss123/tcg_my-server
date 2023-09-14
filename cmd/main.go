package main

import "tcg_my/pkg/pricefinder"

func main() {
	finder := pricefinder.NewBigWebPriceFinder()
	finder.FindPrices("DBVS-JP010")
	finder2 := pricefinder.NewYuyuteiPriceFinder()
	finder2.FindPrices("AC03-JP006")

	select {}
}
