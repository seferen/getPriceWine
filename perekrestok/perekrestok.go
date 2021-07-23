package perekrestok

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"sync"
)

const SheetName = "perekrestok.ru"

type Perekrestok struct {
	categories []string
	token      string
}

func NewPerkrestok() Perekrestok {
	p := Perekrestok{}
	p.categories = []string{"Вино"}
	return p
}

func (p *Perekrestok) getToken() error {
	resp, err := http.DefaultClient.Get("https://perekrestok.ru")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	reg := regexp.MustCompile(`"accessToken":"(.+?)"`)

	p.token = reg.FindStringSubmatch(buf.String())[1]
	log.Println("was found token:", p.token)
	return nil
}

type Category struct {
	Id       int    `json:"id"`
	Title    string `json:"title"`
	Slug     string `json:"slug"`
	ParentId int    `json:"parentId"`
}

type PerekrestokFeedResp struct {
	Content struct {
		Category Category
		Items    []struct {
			Category Category `json:"category"`
			Count    int      `json:"count"`
			Products []struct {
				Id    int    `json:"id"`
				Title string `json:"title"`
			} `json:"products"`
		} `json:"items"`
	} `json:"content"`
}

type PerekrestokProdRequest struct {
	Filter struct {
		Category int `json:"category"`
	} `json:"filter"`
}

type PerekrestocCatResp struct {
	Content struct {
		Categories []struct {
			Category Category `json:"category"`
		} `json:"categories"`
	} `json:"content"`
}

func (p *Perekrestok) GetList(xlsxWriter chan interface{}, wg *sync.WaitGroup) error {

	defer wg.Done()

	err := p.getToken()
	if err != nil {
		return err
	}

	// Get categories of products
	resp, err := p.NewPerekrestokRequest(http.MethodGet, "https://www.perekrestok.ru/api/customer/1.4.0.0/catalog", nil)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Println("perekrestok.ru get categories status:", resp.Status)

	if resp.StatusCode != 200 {
		return errors.New("was receive from perekrestok.ru status: " + resp.Status)
	}

	// decode response from categories request
	dec := json.NewDecoder(resp.Body)

	cat := PerekrestocCatResp{}

	err = dec.Decode(&cat)
	if err != nil {
		return err
	}

	// find parent category
	for _, vCat := range cat.Content.Categories {
		if vCat.Category.Title == "Алкогольные напитки" {
			resp1, err := p.NewPerekrestokRequest(
				http.MethodGet,
				fmt.Sprintf("https://www.perekrestok.ru/api/customer/1.4.0.0/catalog/category/feed/%d", vCat.Category.Id),
				nil)
			if err != nil {
				return err
			}
			defer resp1.Body.Close()
			log.Println("perekrestok get information about categoryes:", resp1.Status)

			feed := PerekrestokFeedResp{}
			dec = json.NewDecoder(resp1.Body)

			err = dec.Decode(&feed)

			for _, vFeed := range feed.Content.Items {
				log.Println("Category:", vFeed.Category.Title, "Id:", vFeed.Category.Id, "Count:", vFeed.Count, "Products count:", len(vFeed.Products))
				reqProd := PerekrestokProdRequest{Filter: struct {
					Category int "json:\"category\""
				}{Category: vFeed.Category.Id}}

				if reqProdB, err := json.Marshal(reqProd); err == nil {
					if respProd, err := p.NewPerekrestokRequest(http.MethodPost, "	https://www.perekrestok.ru/api/customer/1.4.0.0/catalog/search/form", bytes.NewReader(reqProdB)); err == nil {
						defer respProd.Body.Close()

					}

				} else {
					log.Println(err)
				}

			}

		}
	}

	// get information about parrent category

	return nil
}

func (p *Perekrestok) NewPerekrestokRequest(method string, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Auth",
		"Bearer "+p.token)

	log.Println(req)

	return http.DefaultClient.Do(req)

}
