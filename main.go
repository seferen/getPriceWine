package main

import (
	"log"
	"regexp"
	"strconv"
	"sync"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/antchfx/htmlquery"
	"github.com/seferen/getPriceWine/auchan"
	"github.com/seferen/getPriceWine/bristol"
	"github.com/seferen/getPriceWine/defaulttypes"
	"github.com/seferen/getPriceWine/lenta"
	"github.com/seferen/getPriceWine/metro"
	"github.com/seferen/getPriceWine/winelab"
)

var re = regexp.MustCompile(`[\s|a]`)
var abs = []rune("ABCDEFGHIJKLMNOPQRSTUVWXUZ")
var indexes = map[string]int{}

func init() {
	log.Println("Application was started")
	log.Println("alphabe for xlsx file was initiate:", string(abs))
	for _, v := range []string{metro.SheetName, lenta.SheetName, auchan.SheetName, bristol.SheetName, winelab.SheetName} {
		indexes[v] = 2
	}
}

func main() {

	defer func() {
		log.Println("**********was writen***********")
		for k, v := range indexes {
			log.Println(k, v)

		}
		log.Println("Application was finished")
	}()

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

		wg.Add(1)

		go func(wg *sync.WaitGroup) {
			err := metro.GetList(fWr, wg)
			if err != nil {
				log.Println(err)
			}
		}(&wg)

		wg.Add(1)

		go func(wg *sync.WaitGroup) {
			err := lenta.GetList(fWr, wg)

			if err != nil {
				log.Println(err)
			}
		}(&wg)

		wg.Add(1)

		go func() {
			auchan.GetList(fWr, &wg)
		}()

		wg.Add(1)
		go func() {
			b := bristol.NewItem()

			err := b.GetList(fWr, &wg)
			if err != nil {
				log.Println(err)
			}
		}()

		wg.Add(1)
		go func() {
			win := winelab.NewItem()
			if err := win.GetList(fWr, &wg); err != nil {
				log.Println(err)
			}
		}()

		wg.Wait()

		quite <- 0

	}(fWr, qu)

	// processing
	for {
		select {
		case obj := <-fWr:
			switch objType := obj.(type) {

			case defaulttypes.XlsxWriter:
				log.Println("Test xlsx writer")
				objType.XlsxWrite()

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
			case auchan.AuchanProdResp:
				// "id", "category", "name", "price"
				log.Println(len(objType.Items))
				for i, v := range objType.Items {
					ind := strconv.Itoa(i + indexes[auchan.SheetName])

					xlf.SetCellValue(auchan.SheetName, "A"+ind, v.Id)
					xlf.SetCellValue(auchan.SheetName, "B"+ind, v.CategoryCodes[len(v.CategoryCodes)-1].Name)
					xlf.SetCellValue(auchan.SheetName, "C"+ind, v.Title)
					xlf.SetCellValue(auchan.SheetName, "D"+ind, v.Price.Value)

				}
				indexes[auchan.SheetName] += len(objType.Items)
				log.Println("auchan indexes:", indexes[auchan.SheetName])

			case *bristol.BristolResp:
				// "name", "regularPrice", "cardPrice"

				for i, v := range objType.Nodes {

					ind := strconv.Itoa(i + indexes[bristol.SheetName])

					title := htmlquery.InnerText(htmlquery.FindOne(v, `//div[@class="catalog-title"]/a`))
					priceN := htmlquery.FindOne(v, `/div/div/div[@class="catalog-price-block"]/span[contains(@class, "catalog-price") and contains(@class, "new")]`)
					priceO := htmlquery.FindOne(v, `/div/div/div[@class="catalog-price-block"]/span[contains(@class, "catalog-price") and contains(@class, "old")]`)
					priceC := htmlquery.FindOne(v, `/div/div/div[@class="catalog-price-block"]/span[@class="catalog-price"]`)

					xlf.SetCellValue(bristol.SheetName, "A"+ind, title)

					if priceC != nil {
						price := re.ReplaceAllString(htmlquery.InnerText(priceC), "")
						xlf.SetCellValue(bristol.SheetName, "B"+ind, price)
						xlf.SetCellValue(bristol.SheetName, "C"+ind, price)

					} else {
						xlf.SetCellValue(bristol.SheetName, "B"+ind, re.ReplaceAllLiteralString(htmlquery.InnerText(priceN), ""))
						xlf.SetCellValue(bristol.SheetName, "C"+ind, re.ReplaceAllLiteralString(htmlquery.InnerText(priceO), ""))

					}

				}
				indexes[bristol.SheetName] += len(objType.Nodes)
				log.Println("bristol indexes:", indexes[bristol.SheetName])
			case *winelab.WinelabResp:

				// "county", "name", "regularPrice", "withoutDiscont"
				for i, v := range objType.Nodes {
					ind := strconv.Itoa(i + indexes[winelab.SheetName])
					country := htmlquery.InnerText(htmlquery.FindOne(v, `//div[@class="product_card--header"]/div[@class="country_wrapper"]/h3`))
					title := htmlquery.InnerText(htmlquery.FindOne(v, `//div[@class="product_card--header"]/div[3]`))

					xlf.SetCellValue(winelab.SheetName, "A"+ind, country)
					xlf.SetCellValue(winelab.SheetName, "B"+ind, title)

					if discount := htmlquery.FindOne(v, `//div[@class="product_card--footer__container"]/div/div/div[@class="discount"]/span[@class="discount__value"]`); discount != nil {
						xlf.SetCellValue(winelab.SheetName, "C"+ind, defaulttypes.ReDigit.ReplaceAllString(htmlquery.InnerText(discount), ""))

					} else {
						xlf.SetCellValue(winelab.SheetName, "C"+ind, "0")

					}

					if dataPrice := htmlquery.FindOne(v, `//div[@class="product_card_cart"]/div/@data-price`); dataPrice != nil {
						xlf.SetCellValue(winelab.SheetName, "D"+ind, htmlquery.InnerText(dataPrice))

					} else {
						xlf.SetCellValue(winelab.SheetName, "C"+ind, "0")

					}

				}
				indexes[winelab.SheetName] += len(objType.Nodes)
				log.Println("winelab indexes:", indexes[winelab.SheetName])

			}
		case <-qu:
			log.Println("stop working job")
			return
		default:
		}
	}

}
