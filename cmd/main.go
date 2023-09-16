package main

import "github.com/jollyboss123/tcg_my-server/pkg/api"

var Version = "v0.1.0"

func main() {
	s := api.NewServer(api.WithVersion(Version))
	s.Init()
	s.Run()

	//finder := source.NewBigWeb()
	//finder.List(context.Background(), "DBVS-JP010")
	//finder2 := source.NewYYT()
	//f, err := finder2.List(context.Background(), "AC03-JP006")
	//if err != nil {
	//	log.Println(err)
	//}
	//log.Println(f)
}
