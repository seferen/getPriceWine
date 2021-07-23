package metro

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"github.com/seferen/getPriceWine/common"
)

var index int = 2

const sheetName = "metro-cc.ru"

type Metro struct {
	baseUrl    *url.URL
	categories map[string]int
}

func (m *Metro) GetWriten() string {
	return fmt.Sprintf("%s: %d", sheetName, index)
}
func (m *Metro) NewItem() {
	log.Println("Create new instanse")
	var err error
	m.baseUrl, err = url.Parse("https://api.metro-cc.ru")
	if err != nil {
		log.Panic(err)
	}

	m.categories = map[string]int{"Вино": -1, "Шампанское и игристые вина": -1}

}

func (m *Metro) WriteHeader(xlsxWriter chan common.XlsxWriter, wg *sync.WaitGroup) error {
	defer wg.Done()

	// write headers to xlsx file
	xlsxWriter <- &common.HeadSheet{SheetName: sheetName,
		HeadNames: []string{"id", "category", "name", "prices price", "prices oldprice", "pricesPerUnit price", "pricesPerUnit oldPrice"}}

	log.Println("Header was writen")
	return nil
}

type MetroCatResp struct {
	Success bool `json:"success"`
	Data    []struct {
		Id       int    `json:"id"`
		ParentId int    `json:"parent_id"`
		Name     string `json:"name"`
	} `json:"data"`
	Errors []string `json:"errors"`
}

type MetroProdResp struct {
	Success bool `json:"success"`
	Data    struct {
		CurrentPage int `json:"current_page"`
		Data        []struct {
			Id         int    `json:"id"`
			CategoryId int    `json:"category_id"`
			Article    int    `json:"article"`
			Name       string `json:"name"`
			Prices     struct {
				Price    float32 `json:"price"`
				OldPrice float32 `json:"old_price"`
			} `json:"prices"`
			PricesPerUnit struct {
				Price    float32 `json:"price"`
				OldPrice float32 `json:"old_price"`
			} `json:"prices_per_unit"`
		} `json:"data"`
		LastPage int `json:"last_page"`
		To       int `json:"to"`
		Total    int `json:"total"`
	} `json:"data"`
	Errors         []string `json:"errors"`
	CategoryString string
}

func (m *MetroProdResp) XlsxWrite() error {
	// "id", "category", "name", "prices price", "prices oldprice", "pricesPerUnit price", "pricesPerUnit oldPrice"

	log.Println("Current page:", m.Data.CurrentPage, "Last Page:", m.Data.LastPage)

	for i, v := range m.Data.Data {

		ind := strconv.Itoa(i + index)

		common.Xlsx.SetCellValue(sheetName, "A"+ind, v.Id)
		common.Xlsx.SetCellValue(sheetName, "B"+ind, m.CategoryString)
		common.Xlsx.SetCellValue(sheetName, "C"+ind, v.Name)
		common.Xlsx.SetCellValue(sheetName, "D"+ind, v.Prices.Price)
		common.Xlsx.SetCellValue(sheetName, "E"+ind, v.Prices.OldPrice)
		common.Xlsx.SetCellValue(sheetName, "F"+ind, v.PricesPerUnit.Price)
		common.Xlsx.SetCellValue(sheetName, "G"+ind, v.PricesPerUnit.OldPrice)

	}
	index += len(m.Data.Data)
	log.Println("metro write:", index)
	return nil
}

func (m *Metro) GetList(xlsxWriter chan common.XlsxWriter, wg *sync.WaitGroup) error {
	defer wg.Done()

	urlStr := *m.baseUrl

	urlStr.Path = "/api/v1/C98BB1B547ECCC17D8AEBEC7116D6/10/categories/"
	log.Println(urlStr.String())

	// request categories
	resp, err := http.Get("https://api.metro-cc.ru/api/v1/C98BB1B547ECCC17D8AEBEC7116D6/10/categories/")

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// check status code from request of categories
	if resp.StatusCode != 200 {
		return errors.New("Status code from request categories: " + resp.Status)
	}

	//decode request body
	metroCat := MetroCatResp{}
	dec := json.NewDecoder(resp.Body)

	err = dec.Decode(&metroCat)

	if err != nil {
		return err
	}

	// find the name of categories and products and writing it to xlsx file
	for cat := range m.categories {

		for _, v := range metroCat.Data {

			if v.Name == cat {
				ids := []int{}
				for _, subcat := range metroCat.Data {
					if subcat.ParentId == v.Id {
						ids = append(ids, subcat.Id)
					}
				}

				if len(ids) == 0 {
					ids = append(ids, v.Id)
				}

				m.categories[cat] = v.Id
				// log.Println(cat, v)

				// find products and writing it to xlsx file
				err := m.getProducts(ids, v.Name, 100, 1, xlsxWriter)
				if err != nil {
					log.Println(err)
				}
				break
			}
		}
	}

	return nil
}

func (m *Metro) getProducts(category []int, categoryName string, paginate int, page int, xlsxWriter chan common.XlsxWriter) error {
	log.Println("Get products:", category, "paginate:", 100, "page:", page)

	urlStr := *m.baseUrl
	urlStr.Path = fmt.Sprintf("/api/v1/C98BB1B547ECCC17D8AEBEC7116D6/10/products")
	query := url.Values{}
	for i, cat := range category {
		query.Set("category_id["+strconv.Itoa(i)+"]}", strconv.Itoa(cat))
	}

	query.Set("paginate", "100")
	query.Set("page", strconv.Itoa(page))

	urlStr.RawQuery = query.Encode()

	log.Println(urlStr.String())
	resp, err := http.Get(urlStr.String())
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	productResp := MetroProdResp{}
	dec := json.NewDecoder(resp.Body)

	err = dec.Decode(&productResp)
	if err != nil {
		return err
	}

	productResp.CategoryString = categoryName

	xlsxWriter <- &productResp

	if page <= productResp.Data.LastPage {
		m.getProducts(category, categoryName, 100, page+1, xlsxWriter)
	}

	return nil
}
