package auchan

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/seferen/getPriceWine/common"
)

const sheetName = "auchan.ru"

var index int = 2

var categories = []string{"Вино"}

type Auchan struct {
}

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

func (a *AuchanProdResp) XlsxWrite() error {
	// "id", "category", "name", "price"
	// log.Println(len(a.Items))
	for i, v := range a.Items {
		ind := strconv.Itoa(i + index)

		common.Xlsx.SetCellValue(sheetName, "A"+ind, v.Id)
		common.Xlsx.SetCellValue(sheetName, "B"+ind, v.CategoryCodes[len(v.CategoryCodes)-1].Name)
		common.Xlsx.SetCellValue(sheetName, "C"+ind, v.Title)
		common.Xlsx.SetCellValue(sheetName, "D"+ind, v.Price.Value)

	}
	index += len(a.Items)
	log.Println("Auchan write:", index)
	return nil
}

func (a *Auchan) GetWriten() string {
	return fmt.Sprintf("%s: %d", sheetName, index)
}

func (a *Auchan) NewItem() {
	log.Println("Create new instanse")

}

func (a *Auchan) WriteHeader(xlsxWriter chan common.XlsxWriter, wg *sync.WaitGroup) error {
	defer wg.Done()

	xlsxWriter <- &common.HeadSheet{SheetName: sheetName, HeadNames: []string{"id", "category", "name", "price"}}

	log.Println("Header was writen")

	return nil
}

func (a *Auchan) GetList(xlsxWriter chan common.XlsxWriter, wg *sync.WaitGroup) error {
	defer wg.Done()

	log.Println("Start getting data")

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

func findCategories(cat AuchanCategItem, catName string, xlsxWriter chan common.XlsxWriter) error {
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

			xlsxWriter <- &prod
		}

	} else if len(cat.Items) != 0 {
		for _, v := range cat.Items {
			err := findCategories(v, catName, xlsxWriter)
			return err
		}

	}

	return nil
}
