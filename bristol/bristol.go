package bristol

import (
	"log"
	"net/url"
	"strconv"
	"sync"

	"github.com/antchfx/htmlquery"
	"github.com/seferen/getPriceWine/defaulttypes"
	"golang.org/x/net/html"
)

const SheetName string = "bristol.ru"

type Bristol struct {
	Url   string
	Paths []string
}

type BristolResp struct {
	Index int
	Nodes []*html.Node
}

func NewItem() Bristol {

	return Bristol{Url: "bristol.ru", Paths: []string{"/catalog/alkogol/vina_vinogradnye/", "/catalog/alkogol/igristye_vina/", "/catalog/alkogol/vermut_1/"}}

}

func (b *Bristol) GetList(xlsxWriter chan interface{}, wg *sync.WaitGroup) error {
	defer wg.Done()
	// writing header to xlsx file
	xlsxWriter <- defaulttypes.HeadSheet{SheetName: SheetName,
		HeadNames: []string{"name", "regularPrice", "cardPrice"}}

	for _, path := range b.Paths {
		err := b.loadItems(path, 0, 1, xlsxWriter)

		if err != nil {
			log.Println(err)
		}
	}
	return nil
}

func (b *Bristol) loadItems(urlPath string, index int, page int, xlsxWriter chan interface{}) error {
	uri := url.URL{}
	uri.Scheme = "https"
	uri.Host = b.Url
	uri.Path = urlPath

	query := url.Values{}
	query.Set("sort", "price")
	query.Set("order", "asc")
	query.Set("PAGEN_1", strconv.Itoa(page))
	uri.RawQuery = query.Encode()

	log.Println(uri.String())

	nd, err := htmlquery.LoadURL(uri.String())
	if err != nil {
		return err
	}

	totalCount, err := strconv.Atoi(defaulttypes.ReDigit.ReplaceAllLiteralString(
		htmlquery.InnerText(htmlquery.FindOne(nd, `//div[@class="titleRightDesign"]/span`)), ""))

	log.Println("total count:", totalCount, "index:", index)

	ndres := htmlquery.Find(nd, string(`//div[@class="col-sm-3 wrapListItem catalog-box"]`))

	xlsxWriter <- &BristolResp{Index: index, Nodes: ndres}

	if len(ndres)+index < totalCount {
		if err = b.loadItems(urlPath, index+len(ndres), page+1, xlsxWriter); err != nil {
			return err
		}

	}
	// log.Println(ndres)

	return nil
}
