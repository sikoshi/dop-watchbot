package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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

func main()  {

	start := time.Now()

	cityId := "5f5f1e3b4c8a49e692fefd70" // this code of almaty in techodom's registry
	targetUrl := "https://api.technodom.kz/katalog/api/v1/products/category/af-products?city_id=" + cityId +
		"&limit=5000&sorting=score&price=0"

	client := http.Client{}

	req, err := http.NewRequest("GET", targetUrl, nil)

	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Set("affiliation","web")

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
		for _, p := range r.Payload {
			
			break
		}
	}

	fmt.Println(len(r.Payload))
	fmt.Println(r.Total)

	elapsed := time.Since(start)
	log.Printf("time: %s", elapsed)
}
