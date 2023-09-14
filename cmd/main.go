package main

import "github.com/jollyboss123/tcg_my-server/pkg/source"

func main() {
	finder := source.NewBigWeb()
	finder.Scrape("DBVS-JP010")
	finder2 := source.NewYYT()
	finder2.Scrape("AC03-JP006")

	select {}
}
