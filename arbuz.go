package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var DB *sqlx.DB

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
	PriceActual     float64  `json:"priceActual"`
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

		//fmt.Println(url)

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
					//break
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
	storedProduct := new(productType)

	dbErr := DB.QueryRow("SELECT * FROM arbuz_products WHERE id = $1", p.ID).Scan(&storedProduct)

	if dbErr != nil && dbErr == sql.ErrNoRows {

		// insert product data
		//fmt.Printf("%v \n", p)

		_, err := DB.Exec("INSERT INTO arbuz_products (" +
			"id,catalog_id,name,producer_country,brand_name,description,uri,image,measure,is_weighted,weight_avg," +
			"weight_min,weight_max,piece_weight_max,quantity_min_step,barcode,is_available,is_local) " +
			"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)",
			p.ID, p.CatalogID, p.Name, p.ProducerCountry, p.BrandName, p.Description,
			p.URI, p.Image, p.Measure, p.IsWeighted, p.WeightAvg, p.WeightMin, p.WeightMax,
			p.PieceWeightMax, p.QuantityMinStep, p.Barcode, p.IsAvailable, p.IsLocal)
		if err != nil {
			fmt.Println(err)
		}
	} else {
		// update product data
	}

	// products_prices table
	// get last price for this product, if previous price differs from current then insert new one
	lastPrice := 0
	currPrice := int(p.PriceActual)

	pErr := DB.QueryRow("SELECT product_price FROM arbuz_prices WHERE product_id = $1 AND is_last=true", p.ID).Scan(&lastPrice)

	if pErr != nil && pErr == sql.ErrNoRows {
		DB.Exec("INSERT INTO arbuz_prices (product_id, product_price, is_last) VALUES ($1, $2, true)",
			p.ID, currPrice)
	} else if lastPrice != currPrice {

		//fmt.Printf("%d => %d\n", lastPrice, currPrice)

		DB.Exec("UPDATE arbuz_prices SET is_last=false WHERE product_id=$1 AND is_last=true", p.ID)
		DB.Exec("INSERT INTO arbuz_prices (product_id, product_price, is_last) VALUES ($1, $2, true)", p.ID, currPrice)
	}
}

func initDB() {
	
	// := "watchbot_pg:watchbot_pg@127.0.0.1:5432/watchbot_pg"

	var pgSqlConnectionString string

	if val, exists := os.LookupEnv("PGSQLCONNECTIONSTRING"); exists {
		pgSqlConnectionString = val
	} else if err := godotenv.Load(".env"); err == nil {
		pgSqlConnectionString = os.Getenv("PGSQLCONNECTIONSTRING")
	}

	db, err := sqlx.Connect("pgx", "postgres://" + pgSqlConnectionString)
	if err != nil {
		log.Fatalln(err)
	}

	DB = db
}

func main ()  {

	// establishing database connection
	initDB()

	// categories
	c := make(map[int]categoryType)

	// getting jwt token for further api calls
	jwtToken, phpSessId, catalogString := initScrap()

	if err := json.Unmarshal([]byte(catalogString), &c); err != nil {
		// panic
		fmt.Println(err)
	}

	for _, v := range c {

		processCategory(jwtToken, phpSessId, v.Id)

		//break
	}
}