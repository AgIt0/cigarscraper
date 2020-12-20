package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

type Availability struct {
	available bool
	price     float64
	currency  string
}

type Cigar struct {
	name   string
	url    string
	single Availability
	box    Availability
}

func main() {
	fName := "cigars.json"
	file, err := os.Create(fName)
	if err != nil {
		log.Fatalf("Cannot create file %q: %s\n", fName, err)
		return
	}
	defer file.Close()

	c := colly.NewCollector(
		colly.AllowedDomains("www.hacico.de"),
		colly.CacheDir("./hacico_cache"),
	)

	cigars := make([]Cigar, 0, 200)

	c.OnHTML(`a.sub3`, func(e *colly.HTMLElement) {
		// For some reason there are a few "hiddne" pages that just don't have a name in the link
		if len(e.Text) == 0 {
			return
		}
		link := e.Attr("href")

		e.Request.Visit(link)
	})

	c.OnRequest(func(r *colly.Request) {
		log.Println("visiting", r.URL.String())
	})

	c.OnHTML(`div.product_listing_box`, func(e *colly.HTMLElement) {
		name := e.ChildText("a.product_listing_box_name")
		url := e.ChildAttr("a.product_listing_box_name", "href")

		log.Println("Scraping ", name)

		cigar := Cigar{
			name: name,
			url:  url,
		}

		e.ForEach("table table tr", func(i int, el *colly.HTMLElement) {
			var price float64
			available := true

			if el.DOM.Find("td:nth-child(5) > input").Length() == 0 {
				available = false
			}

			priceWithCurrency := strings.Split(strings.TrimSpace(el.ChildText("td:nth-child(3)")), " ")
			currency := priceWithCurrency[0]

			if currency == "" {
				available = false
				price = 0.0
			} else {
				price, _ = strconv.ParseFloat(strings.ReplaceAll(priceWithCurrency[1], ",", "."), 64)
			}

			availability := Availability{
				available: available,
				currency:  currency,
				price:     price,
			}

			if i == 0 {
				cigar.single = availability
			} else {
				cigar.box = availability
			}
		})

		cigars = append(cigars, cigar)
	})

	c.Visit("https://www.hacico.de/en/Cigars/Nicaragua")

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")

	enc.Encode(cigars)
}
