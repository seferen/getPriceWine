package lenta

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/360EntSecGroup-Skylar/excelize"
)

var categories = [...]string{"Вино", "Вермуты", "Десертные вина", "Игристое Вино", "Изысканный выбор"}

const sheetName string = "lenta.ru"

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

func List(cli *http.Client, fileXLSX *excelize.File) {

	resp, err := lentaHttpRequest(cli, http.MethodGet,
		"https://lenta.com/api/v1/stores/0006/catalog", nil)

	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)

	catalog := make([]position, 4)

	log.Println("lenta list result request:", resp.Status)

	err = dec.Decode(&catalog)
	if err != nil {
		log.Panicln(err)
	}
	log.Println("catalog was got:", catalog)

	category := make(map[string]int)

	for _, v := range catalog {
		v.getIds(category)
	}
	var index int
	if len(category) != 0 {
		// Create a new sheet.
		index = fileXLSX.NewSheet(sheetName)
		fileXLSX.SetCellValue(sheetName, "A1", "code")
		fileXLSX.SetCellValue(sheetName, "B1", "name")
		fileXLSX.SetCellValue(sheetName, "C1", "brand")
		fileXLSX.SetCellValue(sheetName, "D1", "regularPrice")
		fileXLSX.SetCellValue(sheetName, "E1", "cardPrice")
		fileXLSX.SetCellValue(sheetName, "F1", "category")
	}

	log.Println(category)

	var i int = 2
	for k, v := range category {
		log.Println(k, v)

		lReq := NewLentaRequest(k, v)
		b, err := json.Marshal(lReq)
		if err != nil {
			log.Println(err)
			continue
		}

		r := bytes.NewReader(b)

		resp, err := lentaHttpRequest(cli, http.MethodPost, "https://lenta.com/api/v1/skus/list", r)

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
		log.Println(lResp)

		lResp.WriteXLS(fileXLSX, i)

		i += len(lResp.Skus)

	}
	fileXLSX.SetActiveSheet(index)

	// defer resp.Body.Close()

	// return lResp

}

func (lResp *LentaResponse) WriteXLS(xlsxFile *excelize.File, startIndex int) {

	stopIndex := len(lResp.Skus) + startIndex
	log.Println("start index:", startIndex, "stop index:", stopIndex, "count:", len(lResp.Skus))

	for i := startIndex; i < stopIndex; i++ {
		v := lResp.Skus[i-startIndex]
		// log.Println(i, v)
		iRow := strconv.Itoa(i)
		xlsxFile.SetCellValue(sheetName, "A"+iRow, v.Code)
		xlsxFile.SetCellValue(sheetName, "B"+iRow, v.Title)
		xlsxFile.SetCellValue(sheetName, "C"+iRow, v.Brand)
		xlsxFile.SetCellValue(sheetName, "D"+iRow, v.RegularPrice.Value)
		xlsxFile.SetCellValue(sheetName, "E"+iRow, v.CardPrice.Value)
		xlsxFile.SetCellValue(sheetName, "F"+iRow, v.GaCategory)
	}

}

func lentaHttpRequest(cli *http.Client, method string, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-type", "application/json")
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")

	log.Println(req)

	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}

	log.Println(resp.Status)

	return resp, nil

}

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
