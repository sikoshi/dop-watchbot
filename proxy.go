package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
)
func main()  {

	targetUrl := "https://free-proxy-list.net"

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



		ta := doc.Find("textarea")

		if len(ta.Nodes) == 1 {

			var wg sync.WaitGroup

			workingProxyChannel := make(chan string)

			if html, err := ta.Html(); err == nil {

				lines := strings.Split(html, "\n")

				var proxyList []string

				if len(lines) > 0 {
					for _, l := range lines {

						l = strings.TrimSpace(l)

						if l == "" {
							continue
						}

						if matched, _ := regexp.Match(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}):(\d{1,5})`, []byte(l)); matched {
							proxyList = append(proxyList, l)
						}
					}
				}


				wg.Add(len(proxyList))

				for _, v := range proxyList {
					go func(s string) {
						// check if proxy is working

						if proxyUrl, err := url.Parse("https://" + s); err == nil {

							fmt.Println(s)

							http.DefaultTransport = &http.Transport{Proxy: http.ProxyURL(proxyUrl)}

							resp, err := http.Get("https://ifconfig.me")
							if err != nil {
								log.Println(err)
							}
							defer resp.Body.Close()

							if resp.StatusCode == 200 {

								if body, err := ioutil.ReadAll(resp.Body); err == nil {
									fmt.Printf("resp: %v\n", string(body))

									workingProxyChannel <- s
								}
							}
						} else {
							fmt.Printf("ERROR: %v\n", err)
						}
					}(v)
				}
			}

			go func() {
				wg.Wait()
				close(workingProxyChannel)
			}()

			var workingProxyList []string

			for p := range workingProxyChannel {
				workingProxyList = append(workingProxyList, p)
			}
		}
	}
}