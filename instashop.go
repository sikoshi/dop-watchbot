package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"strconv"
	"strings"
	"github.com/gosimple/slug"
)

type insStoreType struct {
	Link string
	Title string
	Slug string
}

type insCat struct {
	Id int
	ShopSlug string
	Link string
	Title string
}

type insProduct struct {
	Id int
	ShopId int
	CategoryId int
	CategoryTitle string
	Price int
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

}

func processShopCategory(c insCat)  {

	p := 1

	l := make(map[int]bool)

	breakLoop := false

	for {

		targetUrl := "https://almaty.instashop.kz" + c.Link + "?&ajax=Y&PAGEN_1=" + strconv.Itoa(p)


		fmt.Println(targetUrl)

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
					p.Id, _ = strconv.Atoi(productId)
				}

				if productPrice, e := plink.Attr("data-price"); e {
					p.Price, _ = strconv.Atoi(productPrice)
				}

				p.Brand,_          = plink.Attr("data-brand")
				p.Title,_          = plink.Attr("data-name")
				p.Link,_           = plink.Attr("href")
				p.CategoryTitle, _ = plink.Attr("active-section")
				p.CategoryId       = c.Id

				fmt.Printf("%v\n", p)

				pl = append(pl, p)
			})

			if len(pl) > 0 {
				for _, v := range pl {
					if !l[v.Id] {
						l[v.Id] = true
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
}

func processStore(s insStoreType)  {

	targetUrl := "https://almaty.instashop.kz" + s.Link

	//fmt.Printf("%s\n", targetUrl)

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

		var cl []insCat

		doc.Find(".b-multi-menu > .b-multi-menu__item").Each(func(i int, m *goquery.Selection) {

			v, _ := m.Find(".b-multi-menu__submenu-title").Html()

			// has children categories
			if v != "" {

				m.Find(".b-multi-menu__submenu .b-multi-menu__link").Each(func(j int, sub *goquery.Selection) {
					var c insCat
					if l, e := sub.Attr("href"); e {
						c.Link = l
					}

					c.Title = strings.TrimSpace(sub.Text())

					cl = append(cl, c)
				})

			} else {

				l := m.Find("a")

				var c insCat

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
				processShopCategory(v)
				// temporary
				//break
			}
		}
	}
}

func main()  {
	stores := getStores()

	if len(stores) > 0 {
		for _, store := range stores {
			processStore(store)
			// temporary
			break
		}
	}
}