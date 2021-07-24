package winelab

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"sync"

	"github.com/antchfx/htmlquery"
	"github.com/seferen/getPriceWine/common"
	"golang.org/x/net/html"
)

const SheetName = "winelab.ru"

var index int = 2

type Winelab struct {
	Url   string
	Paths []string
}

type WinelabResp struct {
	Nodes []*html.Node
}

func (w *WinelabResp) XlsxWrite() error {
	// "county", "name", "regularPrice", "withoutDiscont"
	for i, v := range w.Nodes {
		ind := strconv.Itoa(i + index)
		country := htmlquery.InnerText(htmlquery.FindOne(v, `//div[@class="product_card--header"]/div[@class="country_wrapper"]/h3`))
		title := htmlquery.InnerText(htmlquery.FindOne(v, `//div[@class="product_card--header"]/div[3]`))

		common.Xlsx.SetCellValue(SheetName, "A"+ind, country)
		common.Xlsx.SetCellValue(SheetName, "B"+ind, title)

		if discount := htmlquery.FindOne(v, `//div[@class="product_card--footer__container"]/div/div/div[@class="discount"]/span[@class="discount__value"]`); discount != nil {
			common.Xlsx.SetCellValue(SheetName, "C"+ind, common.ReDigit.ReplaceAllString(htmlquery.InnerText(discount), ""))

		} else {
			common.Xlsx.SetCellValue(SheetName, "C"+ind, "0")

		}

		if dataPrice := htmlquery.FindOne(v, `//div[@class="product_card_cart"]/div/@data-price`); dataPrice != nil {
			common.Xlsx.SetCellValue(SheetName, "D"+ind, htmlquery.InnerText(dataPrice))

		} else {
			common.Xlsx.SetCellValue(SheetName, "C"+ind, "0")

		}

	}
	index += len(w.Nodes)
	log.Println("winelab write:", index)

	return nil
}

func (w *Winelab) GetWriten() string {
	return fmt.Sprintf("%s: %d", SheetName, index)
}

func (w *Winelab) NewItem() {
	log.Println("Create new instance")
	w.Url = "winelab.ru"
	w.Paths = []string{"/catalog/vino", "/catalog/shampanskie-i-igristye-vina"}

}

func (w *Winelab) WriteHeader(xlsxWriter chan common.XlsxWriter, wg *sync.WaitGroup) error {
	defer wg.Done()

	xlsxWriter <- &common.HeadSheet{SheetName: SheetName,
		HeadNames: []string{"county", "name", "regularPrice", "withoutDiscont"}}

	log.Println("Header was writen")

	return nil
}

func (w *Winelab) GetList(xlsxWriter chan common.XlsxWriter, wg *sync.WaitGroup) error {
	defer wg.Done()

	for _, path := range w.Paths {
		w.loadItems(path, 0, 1, xlsxWriter)

	}

	return nil
}

func (w *Winelab) loadItems(urlPath string, index int, page int, xlsxWriter chan common.XlsxWriter) error {
	uri := url.URL{}

	uri.Scheme = "https"
	uri.Host = w.Url
	uri.Path = urlPath

	query := url.Values{}
	query.Set("sort", "price-asc")
	query.Set("page", strconv.Itoa(page))

	uri.RawQuery = query.Encode()

	// log.Println(uri.String())
	rootNode, err := htmlquery.LoadURL(uri.String())

	if err != nil {
		return err
	}

	countTotal, err := strconv.Atoi(
		common.ReDigit.ReplaceAllLiteralString(htmlquery.InnerText(htmlquery.FindOne(rootNode, `//div[@class="catalog-products_sort "]/span`)), ""))
	if err != nil {
		return err
	}
	rowNodes := htmlquery.Find(rootNode, `//div[@class=" col-12 col-sm-6 col-md-6 col-lg-4"]`)

	xlsxWriter <- &WinelabResp{Nodes: rowNodes}

	if len(rowNodes)+index < countTotal {
		err = w.loadItems(urlPath, index+len(rowNodes), page+1, xlsxWriter)
	}
	return nil
}
