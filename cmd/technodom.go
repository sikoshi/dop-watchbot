package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)
type Category struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	URI          string    `json:"uri"`
	Source       string    `json:"source"`
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

func processCategory(c Category) []Category {

	if len(c.Children) > 0 {
		for _, k := range c.Children {
			processCategory(k)
		}
	} else {

	}
}

func main () {

	start := time.Now()

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

	if len(r.Items) > 0 {
		for _, v := range r.Items {
			processCategory(v)
		}
	}n

	elapsed := time.Since(start)
	log.Printf("time: %s", elapsed)
}
