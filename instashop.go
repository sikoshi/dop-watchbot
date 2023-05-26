package main

import (
	"database/sql"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"github.com/gosimple/slug"
	"time"
)

var DB *sqlx.DB

type insStoreType struct {
	StoreId int `db: "store_id"`
	Link string
	Title string
	Slug string
}

type insCat struct {
	CategoryId int
	StoreId int
	ShopSlug string
	Link string
	Title string
}

type insProduct struct {
	ProductId int
	StoreId int
	CategoryId int
	CategoryTitle string
	Price int
	Slug string
	Brand string
	Title string
	Link string
}

func getStores() []insStoreType {

	var r []insStoreType

	targetUrl := "https://almaty.instashop.kz/category/supermarket/"

	resp, err := http.Get(targetUrl)
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == 200 {

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		doc.Find(".stores_list_item").Each(func(i int, sl *goquery.Selection) {

			if !sl.HasClass("hide") {

				var s insStoreType

				if link, exists := sl.Find("a").Attr("href"); exists {
					s.Link = link
				}

				if text, err := sl.Find(".b-stores-list__name").Html(); err == nil {
					s.Title = strings.TrimSpace(text)
				}

				s.Slug = slug.Make(s.Title)

				if s.Link != "" && s.Title != "" {
					r = append(r, s)
				}
			}
		})
	}

	return r
}

func processShopProduct(p insProduct)  {

	var storedProduct insProduct

	dbErr := DB.QueryRow(
		"SELECT product_id, store_id, category_id, brand, title, link " +
			"FROM instashop_products WHERE store_id=$1 AND product_id = $2",
		p.StoreId, p.ProductId).Scan(
			&storedProduct.ProductId, &storedProduct.StoreId, &storedProduct.CategoryId,
			&storedProduct.Brand, &storedProduct.Title, &storedProduct.Link)
	if dbErr != nil && dbErr == sql.ErrNoRows {

		_, err := DB.Exec("" +
			"INSERT INTO instashop_products (product_id, store_id, category_id, brand, title, link) " +
			"VALUES ($1, $2, $3, $4, $5, $6)",
			p.ProductId, p.StoreId, p.CategoryId, p.Brand, p.Title, p.Link)
		if err != nil {
			fmt.Println(err)
		}
	}
	// setting price

	lastPrice := 0
	currPrice := int(p.Price)

	pErr := DB.QueryRow("SELECT product_price FROM instashop_prices " +
		"WHERE product_id = $1 AND store_id=$2 AND is_last=true", p.ProductId, p.StoreId).Scan(&lastPrice)

	if pErr != nil && pErr == sql.ErrNoRows {
		DB.Exec("INSERT INTO instashop_prices (product_id, store_id, product_price, is_last) VALUES ($1, $2,$3, true)",
			p.ProductId, p.StoreId, currPrice)
	} else if lastPrice != currPrice {

		//fmt.Printf("%d => %d\n", lastPrice, currPrice)

		DB.Exec("UPDATE instashop_prices SET is_last=false WHERE product_id=$1 AND store_id=$2 AND is_last=true", p.ProductId, p.StoreId)
		DB.Exec("INSERT INTO instashop_prices (product_id, store_id, product_price, is_last) VALUES ($1, $2, true)", p.ProductId, p.StoreId, currPrice)
	}
}

func processShopCategory(c insCat)  {

	c.CategoryId, _ = strconv.Atoi(path.Base(c.Link))

	// products table
	var storedCategory insCat

	dbErr := DB.QueryRow("SELECT category_id, store_id, title, link " +
		"FROM instashop_categories WHERE store_id=$1 AND category_id = $2",
		c.StoreId, c.CategoryId).Scan(&storedCategory.CategoryId, &storedCategory.StoreId, &storedCategory.Link, &storedCategory.Title)

	if dbErr != nil && dbErr == sql.ErrNoRows {

		_, err := DB.Exec("INSERT INTO instashop_categories (category_id,store_id,title,link) " +
			"VALUES ($1, $2, $3, $4) ", c.CategoryId, c.StoreId, c.Title, c.Link)
		if err != nil {
			fmt.Println(err)
		}
	}

	p := 1

	l := make(map[int]bool)

	var productsToProcess []insProduct

	breakLoop := false

	for {

		targetUrl := "https://almaty.instashop.kz" + c.Link + "?&ajax=Y&PAGEN_1=" + strconv.Itoa(p)

		resp, err := http.Get(targetUrl)
		if err != nil {
			log.Println(err)
		}

		defer resp.Body.Close()

		if resp.StatusCode == 200 {

			doc, err := goquery.NewDocumentFromReader(resp.Body)
			if err != nil {
				log.Fatal(err)
			}

			var pl []insProduct

			doc.Find(".b-product-list__item a").Each(func(i int, plink *goquery.Selection) {

				var p insProduct

				if productId, e := plink.Attr("data-product"); e {
					p.ProductId, _ = strconv.Atoi(productId)
				}

				if productPrice, e := plink.Attr("data-price"); e {
					p.Price, _ = strconv.Atoi(productPrice)
				}

				p.Brand,_          = plink.Attr("data-brand")
				p.Title,_          = plink.Attr("data-name")
				p.Link,_           = plink.Attr("href")
				p.CategoryTitle, _ = plink.Attr("active-section")
				p.CategoryId       = c.CategoryId
				p.StoreId          = c.StoreId

				pl = append(pl, p)
			})

			if len(pl) > 0 {
				for _, v := range pl {
					if !l[v.ProductId] {
						l[v.ProductId] = true

						productsToProcess = append(productsToProcess, v)
					} else {
						breakLoop = true
					}
				}
			}
		}

		if breakLoop {
			break
		}

		p++
	}


	if len(productsToProcess) > 0 {
		for _,product := range productsToProcess {
			processShopProduct(product)
		}
	}
}

func processStore(s insStoreType)  {

	// products table
	var storedShop insStoreType

	dbErr := DB.QueryRow("SELECT store_id, slug, title, link FROM instashop_stores WHERE slug = $1",
		s.Slug).Scan(&storedShop.StoreId, &storedShop.Slug, &storedShop.Title, &storedShop.Link)

	if dbErr != nil && dbErr == sql.ErrNoRows {

		lastInsertId := 0

		err := DB.QueryRow("INSERT INTO instashop_stores (slug,title,link) " +
			"VALUES ($1, $2, $3) RETURNING store_id", s.Slug, s.Title, s.Link).Scan(&lastInsertId)
		if err != nil {
			fmt.Println(err)
		}

		s.StoreId = lastInsertId

	} else {
		s.StoreId = storedShop.StoreId
	}

	targetUrl := "https://almaty.instashop.kz" + s.Link

	fmt.Printf("%s\n", targetUrl)

	resp, err := http.Get(targetUrl)
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == 200 {

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		// category list
		var cl []insCat

		doc.Find(".b-multi-menu > .b-multi-menu__item").Each(func(i int, m *goquery.Selection) {

			v, _ := m.Find(".b-multi-menu__submenu-title").Html()

			// has children categories
			if v != "" {

				m.Find(".b-multi-menu__submenu .b-multi-menu__link").Each(func(j int, sub *goquery.Selection) {
					var c insCat
					c.StoreId = s.StoreId
					if l, e := sub.Attr("href"); e {
						c.Link = l
					}

					c.Title = strings.TrimSpace(sub.Text())

					cl = append(cl, c)
				})

			} else {

				l := m.Find("a")

				var c insCat

				c.StoreId = s.StoreId

				if t, e := l.Html(); e == nil {
					c.Title = t
				}

				if l, e := l.Attr("href"); e {
					c.Link = l
				}

				cl = append(cl, c)
			}
		})

		if len(cl) > 0 {
			for _, v := range cl {

				if v.Title != "Скидки" && v.Title != "" && v.Link != "" {
					processShopCategory(v)
				}
			}
		}
	}
}

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

	stores := getStores()

	if len(stores) > 0 {
		for _, store := range stores {
			processStore(store)
		}
	}

	elapsed := time.Since(start)
	log.Printf("time: %s", elapsed)
}