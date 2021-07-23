package winelab

import (
	"log"
	"net/url"
	"strconv"
	"sync"

	"github.com/antchfx/htmlquery"
	"github.com/seferen/getPriceWine/defaulttypes"
	"golang.org/x/net/html"
)

const SheetName = "winelab.ru"

type Winelab struct {
	Url   string
	Paths []string
}

type WinelabResp struct {
	Nodes []*html.Node
}

func NewItem() Winelab {
	return Winelab{Url: "winelab.ru", Paths: []string{"/catalog/vino", "/catalog/shampanskie-i-igristye-vina"}}

}

func (w *Winelab) GetList(xlsxWriter chan interface{}, wg *sync.WaitGroup) error {
	defer wg.Done()

	xlsxWriter <- defaulttypes.HeadSheet{SheetName: SheetName,
		HeadNames: []string{"county", "name", "regularPrice", "withoutDiscont"}}

	for _, path := range w.Paths {
		w.loadItems(path, 0, 1, xlsxWriter)

	}

	return nil
}

func (w *Winelab) loadItems(urlPath string, index int, page int, xlsxWriter chan interface{}) error {
	uri := url.URL{}

	uri.Scheme = "https"
	uri.Host = w.Url
	uri.Path = urlPath

	query := url.Values{}
	query.Set("sort", "price-asc")
	query.Set("page", strconv.Itoa(page))

	uri.RawQuery = query.Encode()

	log.Println(uri.String())
	rootNode, err := htmlquery.LoadURL(uri.String())

	if err != nil {
		return err
	}

	countTotal, err := strconv.Atoi(
		defaulttypes.ReDigit.ReplaceAllLiteralString(htmlquery.InnerText(htmlquery.FindOne(rootNode, `//div[@class="catalog-products_sort "]/span`)), ""))
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
