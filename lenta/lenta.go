package lenta

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/seferen/getPriceWine/common"
)

var index int = 2

var categories = []string{"Вино", "Вермуты", "Десертные вина", "Игристое Вино", "Изысканный выбор"}

const SheetName = "lenta.ru"

type Lenta struct {
}

func (l *Lenta) GetWriten() string {
	return fmt.Sprintf("%s: %d", SheetName, index)
}

func (l *Lenta) NewItem() {
	log.Println("Create new instance")

}

func (l *Lenta) WriteHeader(xlsxWriter chan common.XlsxWriter, wg *sync.WaitGroup) error {
	defer wg.Done()

	xlsxWriter <- &common.HeadSheet{SheetName: SheetName,
		HeadNames: []string{"id", "category", "name", "brand", "regularPrice", "cardPrice"}}
	log.Println("Header was writen")
	return nil
}

type position struct {
	Name       string     `json:"name"`
	Code       string     `json:"code"`
	SkuCount   int        `json:"skuCount"`
	Categories []position `json:"categories"`
}

type LentaRequest struct {
	NodeCode    string `json:"nodeCode"`
	TypeSearch  int    `json:"typeSearch"`
	SortingType string `json:"sortingType"`
	Limit       int    `json:"limit"`
}

func NewLentaRequest(category string, limit int) LentaRequest {

	l := LentaRequest{}
	l.NodeCode = category
	l.TypeSearch = 1
	l.Limit = limit
	l.SortingType = "ByCardPriceAsc"
	return l

}

type LentaResponse struct {
	Skus []struct {
		Code         string `json:"code"`
		Title        string `json:"title"`
		Brand        string `json:"brand"`
		RegularPrice struct {
			Value       float32 `json:"value"`
			IntegerPart string  `json:"integerPart"`
			FactionPart string  `json:"fractionPart"`
		} `json:"regularPrice"`
		CardPrice struct {
			Value       float32 `json:"value"`
			IntegerPart string  `json:"integerPart"`
			FactionPart string  `json:"fractionPart"`
		} `json:"cardPrice"`
		GaCategory string `json:"gaCategory"`
	} `json:"skus"`
	Limit      int `json:"limit"`
	TotalCount int `json:"totalCount"`
}

func (l *LentaResponse) XlsxWrite() error {
	// "id", "category", "name", "brand", "regularPrice", "cardPrice"
	log.Println("from lent was got:", len(l.Skus))

	for i, v := range l.Skus {

		ind := strconv.Itoa(i + index)

		common.Xlsx.SetCellValue(SheetName, "A"+ind, v.Code)
		common.Xlsx.SetCellValue(SheetName, "B"+ind, v.GaCategory)
		common.Xlsx.SetCellValue(SheetName, "C"+ind, v.Title)
		common.Xlsx.SetCellValue(SheetName, "D"+ind, v.Brand)
		common.Xlsx.SetCellValue(SheetName, "E"+ind, v.RegularPrice.Value)
		common.Xlsx.SetCellValue(SheetName, "F"+ind, v.CardPrice.Value)

	}

	index += len(l.Skus)

	log.Println("lenta write:", index)
	return nil
}

func (l *Lenta) GetList(xlsxWriter chan common.XlsxWriter, wg *sync.WaitGroup) error {
	defer wg.Done()

	// Get list of exists categories and count of positions
	resp, err := lentaHttpRequest(http.MethodGet,
		"https://lenta.com/api/v1/stores/0006/catalog", nil)

	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)

	catalog := make([]position, 4)

	log.Println("lenta list result request:", resp.Status)

	// Decode json to object
	err = dec.Decode(&catalog)
	if err != nil {
		return err
	}
	log.Println("catalog was got:", catalog)

	//Create map for result of search categories and found it
	category := make(map[string]int)

	for _, v := range catalog {
		v.getIds(category)
	}

	// Create new Sheet and write head of table

	log.Println(category)

	// Get positions from found categories
	for k, v := range category {
		log.Println(k, v)

		// Request about evety foun category
		lReq := NewLentaRequest(k, v)
		b, err := json.Marshal(lReq)
		if err != nil {
			log.Println(err)
			continue
		}

		resp, err := lentaHttpRequest(http.MethodPost, "https://lenta.com/api/v1/skus/list", bytes.NewReader(b))

		if err != nil {
			log.Println(err)
			continue
		}
		defer resp.Body.Close()
		lResp := LentaResponse{}
		dec := json.NewDecoder(resp.Body)
		err = dec.Decode(&lResp)
		if err != nil {
			log.Println(err)
			continue
		}

		xlsxWriter <- &lResp

	}

	return nil

}

func lentaHttpRequest(method string, url string, body io.Reader) (*http.Response, error) {
	// Create new request
	req, err := http.NewRequest(method, url, body)

	if err != nil {
		return nil, err
	}

	// Add header request
	req.Header.Add("Content-type", "application/json")
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")

	log.Println(req)

	// Send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	log.Println(resp.Status)

	return resp, nil

}

// getIds is a function for find id of categories and count position and add it to map
func (p *position) getIds(category map[string]int) {
	for _, v := range categories {
		if v == p.Name {
			category[p.Code] = p.SkuCount
			continue

		} else if len(p.Categories) != 0 {

			for _, v1 := range p.Categories {
				v1.getIds(category)
			}
		}
	}

}
