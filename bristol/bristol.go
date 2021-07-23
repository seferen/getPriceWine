package bristol

import (
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"sync"

	"github.com/antchfx/htmlquery"
	"github.com/seferen/getPriceWine/common"
	"golang.org/x/net/html"
)

var index int = 2

const sheetName string = "bristol.ru"

var re = regexp.MustCompile(`[\s|a]`)

type Bristol struct {
	Url   string
	Paths []string
}

type BristolResp struct {
	Index int
	Nodes []*html.Node
}

func (b *BristolResp) XlsxWrite() error {
	// "name", "regularPrice", "cardPrice"

	for i, v := range b.Nodes {

		ind := strconv.Itoa(i + index)

		title := htmlquery.InnerText(htmlquery.FindOne(v, `//div[@class="catalog-title"]/a`))
		priceN := htmlquery.FindOne(v, `/div/div/div[@class="catalog-price-block"]/span[contains(@class, "catalog-price") and contains(@class, "new")]`)
		priceO := htmlquery.FindOne(v, `/div/div/div[@class="catalog-price-block"]/span[contains(@class, "catalog-price") and contains(@class, "old")]`)
		priceC := htmlquery.FindOne(v, `/div/div/div[@class="catalog-price-block"]/span[@class="catalog-price"]`)

		common.Xlsx.SetCellValue(sheetName, "A"+ind, title)

		if priceC != nil {
			price := re.ReplaceAllString(htmlquery.InnerText(priceC), "")
			common.Xlsx.SetCellValue(sheetName, "B"+ind, price)
			common.Xlsx.SetCellValue(sheetName, "C"+ind, price)

		} else {
			common.Xlsx.SetCellValue(sheetName, "B"+ind, re.ReplaceAllLiteralString(htmlquery.InnerText(priceN), ""))
			common.Xlsx.SetCellValue(sheetName, "C"+ind, re.ReplaceAllLiteralString(htmlquery.InnerText(priceO), ""))

		}

	}
	index += len(b.Nodes)
	log.Println("bristol write:", index)
	return nil
}

func (b *Bristol) GetWriten() string {
	return fmt.Sprintf("%s: %d", sheetName, index)
}

func (b *Bristol) NewItem() {
	log.Println("Create new instance")
	b.Url = "bristol.ru"
	b.Paths = []string{"/catalog/alkogol/vina_vinogradnye/", "/catalog/alkogol/igristye_vina/", "/catalog/alkogol/vermut_1/"}

}

func (b *Bristol) WriteHeader(xlsxWriter chan common.XlsxWriter, wg *sync.WaitGroup) error {
	defer wg.Done()
	// writing header to xlsx file
	xlsxWriter <- &common.HeadSheet{SheetName: sheetName,
		HeadNames: []string{"name", "regularPrice", "cardPrice"}}
	log.Println("Header was writen")
	return nil

}

func (b *Bristol) GetList(xlsxWriter chan common.XlsxWriter, wg *sync.WaitGroup) error {
	defer wg.Done()

	for _, path := range b.Paths {
		err := b.loadItems(path, 0, 1, xlsxWriter)

		if err != nil {
			log.Println(err)
		}
	}
	return nil
}

func (b *Bristol) loadItems(urlPath string, index int, page int, xlsxWriter chan common.XlsxWriter) error {
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

	totalCount, err := strconv.Atoi(common.ReDigit.ReplaceAllLiteralString(
		htmlquery.InnerText(htmlquery.FindOne(nd, `//div[@class="titleRightDesign"]/span`)), ""))

	log.Println("total count:", totalCount, "index:", index)

	ndres := htmlquery.Find(nd, string(`//div[@class="col-sm-3 wrapListItem catalog-box"]`))

	xlsxWriter <- &BristolResp{Index: index, Nodes: ndres}

	if len(ndres)+index < totalCount {
		if err = b.loadItems(urlPath, index+len(ndres), page+1, xlsxWriter); err != nil {
			return err
		}

	}

	return nil
}
