package main

import (
	"log"
	"strconv"
	"sync"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/seferen/getPriceWine/defaulttypes"
	"github.com/seferen/getPriceWine/lenta"
	"github.com/seferen/getPriceWine/metro"
)

var abs = []rune("ABCDEFGHIJKLMNOPQRSTUVWXUZ")
var indexes = map[string]int{}

func init() {
	log.Println("Application was started")
	log.Println("alphabe for xlsx file was initiate:", string(abs))
	indexes[metro.SheetName] = 2
	indexes[lenta.SheetName] = 2
}

func main() {

	defer log.Println("Application wsa finished")

	// create channel for writing to xlsx file information
	fWr := make(chan interface{})
	qu := make(chan int)

	// // create file for writing information
	xlf := excelize.NewFile()
	defer func(file *excelize.File) {
		if err := xlf.SaveAs("result.xlsx"); err != nil {
			log.Println(err)
		}
	}(xlf)

	go func(xlsxWriter chan interface{}, quite chan int) {
		wg := sync.WaitGroup{}

		wg.Add(2)

		go func(wg *sync.WaitGroup) {
			err := metro.GetList(fWr, wg)
			if err != nil {
				log.Println(err)
			}
		}(&wg)

		go func(wg *sync.WaitGroup) {
			err := lenta.GetList(fWr, wg)

			if err != nil {
				log.Println(err)
			}
		}(&wg)

		wg.Wait()

		quite <- 0

	}(fWr, qu)

	// processing
	for {
		select {
		case obj := <-fWr:
			switch objType := obj.(type) {

			case defaulttypes.HeadSheet:

				// create a new sheat
				xlf.NewSheet(objType.SheetName)

				// write headers of sheat
				for i, v := range objType.HeadNames {
					xlf.SetCellValue(objType.SheetName, string(abs[i])+"1", v)

				}
			case metro.MetroProdResp:
				// "id", "category", "name", "prices price", "prices oldprice", "pricesPerUnit price", "pricesPerUnit oldPrice"

				log.Println("Current page:", objType.Data.CurrentPage, "Last Page:", objType.Data.LastPage)

				for i, v := range objType.Data.Data {

					ind := strconv.Itoa(i + indexes[metro.SheetName])

					xlf.SetCellValue(metro.SheetName, "A"+ind, v.Id)
					xlf.SetCellValue(metro.SheetName, "B"+ind, objType.CategoryString)
					xlf.SetCellValue(metro.SheetName, "C"+ind, v.Name)
					xlf.SetCellValue(metro.SheetName, "D"+ind, v.Prices.Price)
					xlf.SetCellValue(metro.SheetName, "E"+ind, v.Prices.OldPrice)
					xlf.SetCellValue(metro.SheetName, "F"+ind, v.PricesPerUnit.Price)
					xlf.SetCellValue(metro.SheetName, "G"+ind, v.PricesPerUnit.OldPrice)

				}
				indexes[metro.SheetName] += len(objType.Data.Data)
				log.Println("metro indexses:", indexes[metro.SheetName])
			case lenta.LentaResponse:
				// "id", "category", "name", "brand", "regularPrice", "cardPrice"
				log.Println("from lent was got:", len(objType.Skus))

				for i, v := range objType.Skus {

					ind := strconv.Itoa(i + indexes[lenta.SheetName])

					xlf.SetCellValue(lenta.SheetName, "A"+ind, v.Code)
					xlf.SetCellValue(lenta.SheetName, "B"+ind, v.GaCategory)
					xlf.SetCellValue(lenta.SheetName, "C"+ind, v.Title)
					xlf.SetCellValue(lenta.SheetName, "D"+ind, v.Brand)
					xlf.SetCellValue(lenta.SheetName, "E"+ind, v.RegularPrice.Value)
					xlf.SetCellValue(lenta.SheetName, "F"+ind, v.CardPrice.Value)

				}

				indexes[lenta.SheetName] += len(objType.Skus)

				log.Println("lenta indexes:", indexes[lenta.SheetName])

			}
		case <-qu:
			log.Println("stop working job")
			return
		default:
		}
	}

}
