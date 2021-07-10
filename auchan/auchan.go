package auchan

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/seferen/getPriceWine/defaulttypes"
)

const SheetName = "auchan.ru"

var categories = []string{"Вино"}

type AuchanCategItem struct {
	Id                  int               `json:"id"`
	Name                string            `json:"name"`
	Code                string            `json:"code"`
	ShowCode            string            `json:"show_code"`
	ProductsCount       int               `json:"productsCount"`
	ActiveProductsCount int               `json:"activeProductsCount"`
	Items               []AuchanCategItem `json:"items"`
}

type AuchanProdResp struct {
	Items []struct {
		Id            int    `json:"id"`
		Title         string `json:"title"`
		CategoryCodes []struct {
			Id   int    `json:"id"`
			Name string `json:"name"`
			Code string `json:"vino"`
		} `json:"categoryCodes"`
		Price struct {
			Value    float32 `json:"value"`
			Currency string  `json:"currency"`
		} `json:"price"`
	} `json:"items"`
}

func GetList(xlsxWriter chan interface{}, wg *sync.WaitGroup) error {
	defer wg.Done()

	// get categories
	resp, err := http.Get("https://www.auchan.ru/v1/categories?node_code=vino&merchant_id=1&active_only=0&show_hidden=1")
	if err != nil {
		return err

	}
	defer resp.Body.Close()

	// decode body
	dec := json.NewDecoder(resp.Body)

	cat := make([]AuchanCategItem, 1)

	err = dec.Decode(&cat)

	if err != nil {
		return err
	}

	xlsxWriter <- defaulttypes.HeadSheet{SheetName: SheetName, HeadNames: []string{"id", "category", "name", "price"}}

	for _, v := range categories {
		for _, c := range cat {
			err := findCategories(c, v, xlsxWriter)
			if err != nil {
				log.Println(err)
			}

		}

	}

	return nil
}

func findCategories(cat AuchanCategItem, catName string, xlsxWriter chan interface{}) error {
	if catName == cat.Name {
		cat.Items = nil
		var maxPage int = cat.ActiveProductsCount/100 + 1

		for i := 1; i <= maxPage; i++ {

			resp, err := http.Get(fmt.Sprintf("https://www.auchan.ru/v1/catalog/products?merchantId=1&filter[category]=%s&filter[active_only]=1&filter[cashback_only]=0&page=%d&perPage=100&orderField=price&orderDirection=asc", cat.Code, i))

			if err != nil {
				return nil
			}
			defer resp.Body.Close()

			dec := json.NewDecoder(resp.Body)

			prod := AuchanProdResp{}

			err = dec.Decode(&prod)

			if err != nil {
				return err
			}

			xlsxWriter <- prod
		}

	} else if len(cat.Items) != 0 {
		for _, v := range cat.Items {
			err := findCategories(v, catName, xlsxWriter)
			return err
		}

	}

	return nil
}
