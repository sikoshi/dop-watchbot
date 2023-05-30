package main

import (
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type apiResponseType struct {
	Payload []struct {
		Sku       string `json:"sku"`
		Title     string `json:"title"`
		Price     string `json:"price"`
		PriceUsd  string `json:"price_usd"`
		OldPrice  string `json:"old_price"`
		Discount  string `json:"discount"`
		Type      string `json:"type"`
		BrandInfo struct {
			Title string `json:"title"`
			Code  string `json:"code"`
		} `json:"brand_info"`
		Brand          string   `json:"brand"`
		BrandCode      string   `json:"brand_code"`
		Rating         int      `json:"rating"`
		Reviews        int      `json:"reviews"`
		CashbackAmount string   `json:"cashback_amount"`
		Score          int      `json:"score"`
		Images         []string `json:"images"`
		Categories     []string `json:"categories"`
		CategoriesRu   []string `json:"categories_ru"`
		CategoriesKz   []string	`json:"categories_kz"`
		Stickers       []struct {
			Title           string    `json:"title"`
			ID              string    `json:"id"`
			Slug            string    `json:"slug"`
			TextColor       string    `json:"text_color"`
			BackgroundColor string    `json:"background_color"`
			Priority        int       `json:"priority"`
			CreatedAt       time.Time `json:"created_at"`
			UpdatedAt       time.Time `json:"updated_at"`
		} `json:"stickers"`
		IsNew          bool `json:"is_new"`
		Showcase       bool `json:"showcase"`
		IsPreorder     bool `json:"is_preorder"`
		OnlyPrepayment bool `json:"only_prepayment"`
		Color          struct {
			Code      string `json:"code"`
			Title     string `json:"title"`
			ColorType int    `json:"color_type"`
			Hex       string `json:"hex"`
		} `json:"color"`
		URI      string `json:"uri"`
		Merchant struct {
			Code string `json:"code"`
			Name string `json:"name"`
		} `json:"merchant"`
		DeliveryInfo struct {
			DeliveryMinDate string `json:"delivery_min_date"`
			DeliveryMaxDate string `json:"delivery_max_date"`
			PickupDate      string `json:"pickup_date"`
		} `json:"delivery_info"`
		UnitMeasurement struct {
			Code    string `json:"code"`
			Name    string `json:"name"`
			MinStep string `json:"min_step"`
		} `json:"unit_measurement"`
	} `json:"payload"`
	MetaData struct {
		MetaTitle       string `json:"meta_title"`
		MetaDescription string `json:"meta_description"`
		MetaHeader      string `json:"meta_header"`
		SeoText         string `json:"seo_text"`
	} `json:"meta_data"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

var DB *sqlx.DB

func insInitDB() {

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

func main()  {

	start := time.Now()


	insInitDB()

	cityId := "5f5f1e3b4c8a49e692fefd70" // this code of almaty in techodom's registry




	// products map
	type rawProduct struct {
		Sku					string `db:"sku"`
		Title				string `db:"title"`
		Brand				string `db:"brand"`
		Uri					string `db:"uri"`
		MerchantCode		string `db:"merchant_code"`
		MerchantName		string `db:"merchant_name"`
		MeasurementCode		string `db:"measurement_code"`
		MeasurementName		string `db:"measurement_name"`
		MeasurementStep		string `db:"measurement_step"`
	}

	var rawProducts []rawProduct
	productsMap := make(map[string]rawProduct)

	if err := DB.Select(&rawProducts,"SELECT sku,title,brand,uri,merchant_code," +
		"merchant_name,measurement_code,measurement_name,measurement_step FROM airba_fresh_products"); err == nil {

		if len(rawProducts) > 0 {
			for _, p := range rawProducts {
				productsMap[p.Sku] = p
			}
		}
	}

	// prices map
	type rawPrice struct {
		ProductSku		string	`db:"product_sku"`
		ProductPrice	int		`db:"product_price"`
	}

	var rawPrices []rawPrice
	pricesMap := make(map[string]int)
	if err := DB.Select(&rawPrices,
		"SELECT product_sku, product_price FROM airba_fresh_prices WHERE is_last=true"); err == nil {

		if len(rawPrices) > 0 {
			for _, pp := range rawPrices {

				pricesMap[pp.ProductSku] = pp.ProductPrice
			}
		}
	}

	fmt.Println("productsMap: ", len(productsMap))
	fmt.Println("pricesMap: ", len(pricesMap))


	page  := 1
	limit := 5000
	count := 0

	breakPagingLoop := false

	for {

		targetUrl := "https://api.technodom.kz/katalog/api/v1/products/category/af-products?city_id=" + cityId +
			"&page=" + strconv.Itoa(page) + "&limit=" + strconv.Itoa(limit) + "&sorting=score&price=0"

		client := http.Client{}

		req, err := http.NewRequest("GET", targetUrl, nil)

		if err != nil {
			fmt.Println(err)
			return
		}

		req.Header.Set("affiliation","web")

		fmt.Println(targetUrl)
		resp, respErr := client.Do(req)

		if respErr != nil {
			fmt.Println(respErr)
			return
		}

		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			fmt.Println(resp.StatusCode)
			return
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			return
		}

		var r apiResponseType

		jsonErr := json.Unmarshal([]byte(string(body)), &r)

		if jsonErr != nil {
			fmt.Println(jsonErr)
			return
		}

		if len(r.Payload) > 0 {

			fmt.Println("count: ", count)

			for _, p := range r.Payload {

				currPrice, err := strconv.Atoi(p.Price)

				if err != nil {
					continue
				}

				// insert product
				if _, ok := productsMap[p.Sku]; !ok {

					_, err := DB.Exec(
						"INSERT INTO airba_fresh_products (sku,title,brand,uri,merchant_code,merchant_name,measurement_code,measurement_name,measurement_step) " +
							"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) ",
						p.Sku, p.Title, p.Brand, p.URI, p.Merchant.Code, p.Merchant.Name, p.UnitMeasurement.Name, p.UnitMeasurement.Name, p.UnitMeasurement.MinStep)
					if err != nil {
						fmt.Println(err)
					}
				}

				// update price
				if lastPrice, ok := pricesMap[p.Sku]; !ok {
					// insert
					if _, err := DB.Exec("INSERT INTO airba_fresh_prices (product_sku, product_price, is_last)" +
						" VALUES ($1, $2, true)", p.Sku, p.Price); err != nil {
						fmt.Println(err)
					}
				} else if lastPrice != currPrice {
					// update
					DB.Exec("UPDATE airba_fresh_prices SET is_last=false WHERE product_sku=$1 AND is_last=true", p.Sku)
					DB.Exec("INSERT INTO airba_fresh_prices (product_sku, product_price, is_last) " +
						"VALUES ($1, $2, true)", p.Sku, currPrice)
				}

				count++

				if count == r.Total {
					breakPagingLoop = true
				}
			}
		}

		page++

		if breakPagingLoop {
			break
		}
	}

	elapsed := time.Since(start)
	log.Printf("time: %s", elapsed)
}
