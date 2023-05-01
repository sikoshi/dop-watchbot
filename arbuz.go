package main

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type categoryType struct {
	Id int `json:"id"`
	Name string `json:"name"`
	Uri string `json:"uri"`
	IconSvg string `json:"iconSvg"`
	IconPng string `json:"iconPng"`
	CssClass string `json:"cssClass"`
	Children map[int]categoryType `json:"children"`
}

type productType struct {
	ID              string  `json:"id"`
	CatalogID       string  `json:"catalogId"`
	Lang            string  `json:"lang"`
	Name            string  `json:"name"`
	ProducerCountry string  `json:"producerCountry"`
	BrandName       string  `json:"brandName"`
	Description     string  `json:"description"`
	URI             string  `json:"uri"`
	Image           string  `json:"image"`
	Measure         string  `json:"measure"`
	IsWeighted      bool    `json:"isWeighted"`
	WeightAvg       float64 `json:"weightAvg"`
	WeightMin       float64 `json:"weightMin"`
	WeightMax       float64	`json:"weightMax"`
	PieceWeightMax  float64 `json:"pieceWeightMax"`
	QuantityMinStep float64 `json:"quantityMinStep"`
	PriceActual     int     `json:"priceActual"`
	Barcode         string  `json:"barcode"`
	IsAvailable     bool    `json:"isAvailable"`
	IsLocal			bool   `json:"isLocal"`
}

type apiResponseType struct {
	Data struct {
		Catalogs struct {
			Data []struct {
				ID           string `json:"id"`
				ParentID     string `json:"parentId"`
				Name         string `json:"name"`
				URI          string `json:"uri"`
				CatalogCount int    `json:"catalogCount"`
				Sort         int    `json:"sort"`
				IsActive     int    `json:"isActive"`
				IconSvg      string    `json:"iconSvg"`
				IconPng      string `json:"iconPng"`
			} `json:"data"`
		} `json:"catalogs"`
		Products struct {
			Data []productType `json:"data"`
			Page struct {
				Current  int `json:"current"`
				Last     int `json:"last"`
				First    int `json:"first"`
				Next     int `json:"next"`
				Previous int `json:"previous"`
				Limit    int `json:"limit"`
				Count    int `json:"count"`
			} `json:"page"`
			Sort []struct {
				Name     string `json:"name"`
				Slug     string `json:"slug"`
				Selected int    `json:"selected"`
				Values   []struct {
					Name  string `json:"name"`
					Value int    `json:"value"`
					Label string    `json:"label"`
				} `json:"values"`
			} `json:"sort"`

			Count int `json:"count"`
		} `json:"products"`
	} `json:"data"`
}

func initScrap () (jwtToken string, phpSessId string, catalogString string) {

	targetUrl := "https://arbuz.kz"

	resp, err := http.Get(targetUrl)
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == 200 {

		for _, cookie := range resp.Cookies() {
			if cookie.Name == "PHPSESSID" {
				phpSessId = cookie.Value
			}
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		doc.Find(".container.mt-4").Each(func(i int, s *goquery.Selection) {

			catalogString, _ = s.Find("div[is=catalog-slider]").Attr(":catalogs")
		})
	}

	//container mt-4

	if phpSessId != "" {
		url := "https://arbuz.kz/api/v1/auth/token"
		method := "POST"

		payload := strings.NewReader(`{"consumer":"arbuz-kz.web.desktop","key":"xtKJnogmEeecoKSkTFi5uqu0tLai7AQm"}`)

		client := &http.Client {}
		req, err := http.NewRequest(method, url, payload)

		if err != nil {
			fmt.Println(err)
			return
		}

		req.Header.Add("authority", "arbuz.kz")
		req.Header.Add("accept", "application/json, text/plain, */*'")
		req.Header.Add("accept-language", "ru-KZ,ru;q=0.9")
		req.Header.Add("content-type", "application/json")
		req.Header.Add("cookie", "PHPSESSID=" + phpSessId)
		req.Header.Add("origin", "https://arbuz.kz")
		req.Header.Add("referer", "https://arbuz.kz/")
		req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36'")

		res, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err)
			return
		}

		var j map[string]map[string]interface{}

		if err := json.Unmarshal(body, &j); err != nil {
			// panic
			fmt.Println(err)
		}

		jwtToken = j["data"]["token"].(string)
	}

	return
}

func processCategory(jwtToken string, phpSessId string, categoryId int)  {

	currPage := 1
	maxPage := 1

	for {

		url := "https://arbuz.kz/api/v1/shop/catalog/" + strconv.Itoa(categoryId) +
			"&page=" + strconv.Itoa(currPage) +
			"&limit=100"

		fmt.Println(url)

		method := "GET"

		client := &http.Client {}

		req, err := http.NewRequest(method, url, nil)

		req.Header.Add("Cookie", "PHPSESSID=" + phpSessId + "; arbuz-kz_jwt_v3=" + jwtToken)
		if err != nil {
			fmt.Println(err)
			return
		}

		res, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			return
		}

		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err)
			return
		}

		var r apiResponseType

		if err := json.Unmarshal([]byte(string(body)), &r); err == nil {

			if len(r.Data.Products.Data) > 0 {
				for _, p := range r.Data.Products.Data {
					processProduct(p)
					break
				}
			}

			maxPage = r.Data.Products.Page.Last

		} else {
			fmt.Println(err)
		}

		if currPage == maxPage {
			break
		} else {
			currPage++
		}
	}
}

func processProduct(p productType) {
	// products table

	// products_prices table

	fmt.Printf("%v\n", p)

	if p.IsAvailable {

	}

	//p.PriceActual
}

func main ()  {

	c := make(map[int]categoryType)

	jwtToken, phpSessId, catalogString := initScrap()

	if err := json.Unmarshal([]byte(catalogString), &c); err != nil {
		// panic
		fmt.Println(err)
	}

	for _, v := range c {

		processCategory(jwtToken, phpSessId, v.Id)

		break
	}
}