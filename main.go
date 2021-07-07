package main

import (
	"log"
	"net/http"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/seferen/getPriceWine/lenta"
)

type WineReciver interface {
	GetList()
	WriteXLS()
}

func main() {

	cli := http.DefaultClient

	log.Println("Application was started")
	xlf := excelize.NewFile()

	lenta.List(cli, xlf)

	if err := xlf.SaveAs("result.xlsx"); err != nil {
		log.Println(err)
	}
	log.Println("Application wsa finished")
}
