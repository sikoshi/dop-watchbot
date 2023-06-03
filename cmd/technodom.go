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

var DB *sqlx.DB

type Category struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	URI          string    `json:"uri"`
	Filters      []struct {
		Code string `json:"code"`
		Title string `json:"title"`
		Value string `json:"value"`
	} `json:"filters"`
	Parent       string    `json:"parent"`
	ItemClass    string    `json:"item_class"`
	CategoryCode string    `json:"category_code"`
	Priority     int       `json:"priority"`
	IsActive     bool      `json:"is_active"`
	Type         string    `json:"type"`
	LinkType     string    `json:"link_type"`
	Children     []Category `json:"children"`
	Icons        struct {
	} `json:"icons"`
	Merchant struct {
		Code string `json:"code"`
		Name string `json:"name"`
	} `json:"merchant"`
	BackgroundColor string `json:"background_color"`
	TextColor       string `json:"text_color"`
}

type Categories struct {
	ID      string `json:"id"`
	HumanID string `json:"human_id"`
	Items   []Category `json:"items"`
}


type CategoryResponse struct {
	Payload []struct {
		Sku       string `json:"sku"`
		Title     string `json:"title"`
		Price     string `json:"price"`
		Type      string `json:"type"`
		Brand          string   `json:"brand"`
		Images         []string `json:"images"`
		Categories     []string `json:"categories"`
		CategoriesRu   []string `json:"categories_ru"`
		CategoriesKz   []string	`json:"categories_kz"`
		URI      string `json:"uri"`
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

func getFinalCategories(c Category) []Category {

	var r []Category

	if len(c.Children) > 0 {
		for _, sub := range c.Children {
			r = append(r, getFinalCategories(sub)...)
		}
	} else {
		r = append(r, c)
	}

	return r
}

func main () {

	start := time.Now()

	initDB()

	cityId := "5f5f1e3b4c8a49e692fefd70"

	targetUrl := "https://api.technodom.kz/menu/api/v1/menu/katalog?city_id=" + cityId

	resp, err := http.Get(targetUrl)
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	var r Categories

	jsonErr := json.Unmarshal([]byte(string(body)), &r)

	if jsonErr != nil {
		fmt.Println(jsonErr)
		return
	}

	var finalCategories []Category

	if len(r.Items) > 0 {
		for _, v := range r.Items {
			finalCategories = append(finalCategories, getFinalCategories(v)...)
		}
	}


	if len(finalCategories) == 0 {
		return
	}

	for _, fc := range finalCategories {

		if len(fc.Filters) > 0 {
			continue
		}

		page  := 1
		limit := 5000
		count := 0

		breakPagingLoop := false

		for {

			fmt.Println(fc)

			targetUrl := "https://api.technodom.kz/katalog/api/v1/products/category/" + fc.CategoryCode + "?city_id=" + cityId +
				"&page=" + strconv.Itoa(page) + "&limit=" + strconv.Itoa(limit) + "&sorting=score&price=0"

			fmt.Println(targetUrl)

			client := http.Client{}

			req, err := http.NewRequest("GET", targetUrl, nil)

			if err != nil {
				fmt.Println(err)
				break
			}

			req.Header.Set("affiliation","web")

			resp, respErr := client.Do(req)

			if respErr != nil {
				fmt.Println(respErr)
				break
			}

			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				fmt.Println(resp.StatusCode)
				break
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println(err)
				break
			}

			var r CategoryResponse

			jsonErr := json.Unmarshal([]byte(string(body)), &r)

			if jsonErr != nil {
				fmt.Println(jsonErr)
				break
			}

			if len(r.Payload) == 0 {
				break
			}


			// products map
			type rawProduct struct {
				Sku					string `db:"sku"`
				Title				string `db:"title"`
				Brand				string `db:"brand"`
				Uri					string `db:"uri"`
			}

			var rawProducts []rawProduct
			productsMap := make(map[string]rawProduct)

			if err := DB.Select(&rawProducts,"SELECT sku,title,brand,uri FROM technodom_products"); err == nil {
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
				"SELECT product_sku, product_price FROM technodom_prices WHERE is_last=true"); err == nil {

				if len(rawPrices) > 0 {
					for _, pp := range rawPrices {

						pricesMap[pp.ProductSku] = pp.ProductPrice
					}
				}
			}

			// products loop
			for _, p := range r.Payload {

				currPrice, err := strconv.Atoi(p.Price)

				if err != nil {
					continue
				}

				// Insert product
				if _, ok := productsMap[p.Sku]; !ok {

					_, err := DB.Exec("INSERT INTO technodom_products (sku,title,brand,uri) " +
						"VALUES ($1, $2, $3, $4) ", p.Sku, p.Title, p.Brand, p.URI)
					if err != nil {
						fmt.Println(err)
					}
				}

				// Update price
				if lastPrice, ok := pricesMap[p.Sku]; !ok {
					// insert
					if _, err := DB.Exec("INSERT INTO technodom_prices (product_sku, product_price, is_last)" +
						" VALUES ($1, $2, true)", p.Sku, p.Price); err != nil {
						fmt.Println(err)
					}
				} else if lastPrice != currPrice {
					// update
					DB.Exec("UPDATE technodom_prices SET is_last=false WHERE product_sku=$1 AND is_last=true", p.Sku)
					DB.Exec("INSERT INTO technodom_prices (product_sku, product_price, is_last) " +
						"VALUES ($1, $2, true)", p.Sku, currPrice)
				}

				count++

				if count >= r.Total {
					breakPagingLoop = true
				}
			}

			page++

			if breakPagingLoop {
				break
			}
		}
	}

	elapsed := time.Since(start)
	log.Printf("time: %s", elapsed)
}
