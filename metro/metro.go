package metro

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/seferen/getPriceWine/defaulttypes"
)

const SheetName = "metro-cc.ru"

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
	} `json:"data"`
	Errors         []string `json:"errors"`
	CategoryString string
}

var categories = map[string]int{"Вино": -1, "Шампанское и игристые вина": -1}

func GetList(xlsxWriter chan interface{}, wg *sync.WaitGroup) error {
	defer wg.Done()
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

	// write headers to xlsx file
	xlsxWriter <- defaulttypes.HeadSheet{SheetName: SheetName,
		HeadNames: []string{"id", "category", "name", "prices price", "prices oldprice", "pricesPerUnit price", "pricesPerUnit oldPrice"}}

	//decode request body
	metroCat := MetroCatResp{}
	dec := json.NewDecoder(resp.Body)

	err = dec.Decode(&metroCat)

	if err != nil {
		return err
	}

	// find the name of categories and products and writing it to xlsx file
	for cat := range categories {

		for _, v := range metroCat.Data {

			if v.Name == cat {
				categories[cat] = v.Id
				log.Println(cat, v)

				// find products and writing it to xlsx file
				err := getProducts(v.Id, v.Name, 100, 1, xlsxWriter)
				if err != nil {
					log.Println(err)
				}
				break
			}
		}
	}

	return nil
}

func getProducts(category int, categoryName string, paginate int, page int, xlsxWriter chan interface{}) error {
	log.Println("Get products:", category, "paginate:", 100, "page:", page)
	resp, err := http.Get(fmt.Sprintf("https://api.metro-cc.ru/api/v1/C98BB1B547ECCC17D8AEBEC7116D6/10/products?category_id[0]=%d&paginate=100&page=%d", category, page))
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

	xlsxWriter <- productResp
	// log.Println(productResp)

	if page <= productResp.Data.LastPage {
		getProducts(category, categoryName, 100, page+1, xlsxWriter)
	}

	return nil
}
