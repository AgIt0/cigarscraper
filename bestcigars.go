package main

import (
	"log"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

func scrapeBestCigars() []Cigar {
	c := colly.NewCollector(
		colly.AllowedDomains("bestcigars.bg"),
		colly.CacheDir("./bestcigars_cache"),
	)

	cigars := make([]Cigar, 0, 200)

	c.OnHTML(`.page-numbers > li > a.next`, func(e *colly.HTMLElement) {
		c.Visit(e.Attr("href"))
	})

	c.OnHTML(`.products, .product-title > a`, func(e *colly.HTMLElement) {
		c.Visit(e.Attr("href"))
	})

	c.OnRequest(func(r *colly.Request) {
		log.Println("visiting", r.URL.String())
	})

	c.OnHTML(`.shop-container > .product`, func(e *colly.HTMLElement) {
		origin := e.DOM.Find(".woocommerce-product-attributes-item--attribute_pa_origin > td > p > a").First().Text()
		name := e.ChildText(".product-title.entry-title")
		url := e.Request.URL.String()

		log.Println("Scraping ", name)

		cigar := Cigar{
			site:   "bestcigars",
			origin: origin,
			name:   name,
			url:    url,
		}

		priceSingleBox := strings.Split(strings.ReplaceAll(strings.TrimSpace(e.DOM.Find(".product-page-price").First().Text()), " лв.", ""), " –")
		priceSingle, _ := strconv.ParseFloat(priceSingleBox[0], 64)

		cigar.single = Availability{
			available: true,
			currency:  "BGN",
			price:     priceSingle,
		}

		cigar.box = Availability{
			available: false,
			currency:  "BGN",
			price:     0.0,
		}

		if len(priceSingleBox) > 1 {
			priceBox, _ := strconv.ParseFloat(strings.TrimSpace(priceSingleBox[1]), 64)
			cigar.box.available = true
			cigar.box.price = priceBox
		}

		cigars = append(cigars, cigar)
	})

	c.Visit("https://bestcigars.bg/kategoriya/puri")

	return cigars
}
