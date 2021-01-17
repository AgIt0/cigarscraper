package main

import (
	"log"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

func getCigarAvailability(el *colly.HTMLElement) Availability {
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

	return Availability{
		available: available,
		currency:  currency,
		price:     price,
	}
}

func scrapeHacico() []Cigar {
	c := colly.NewCollector(
		colly.AllowedDomains("www.hacico.de"),
		colly.CacheDir("./hacico_cache"),
		colly.Async(),
	)

	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 8})
	cigars := make([]Cigar, 0, 200)

	c.OnHTML(`a.sub2`, func(e *colly.HTMLElement) {
		c.Visit(e.Attr("href"))
	})

	c.OnHTML(`a.sub3`, func(e *colly.HTMLElement) {
		// For some reason there are a few "hiddne" pages that just don't have a name in the link
		if len(e.Text) == 0 {
			return
		}
		link := e.Attr("href")

		c.Visit(link)
	})

	c.OnRequest(func(r *colly.Request) {
		log.Println("visiting", r.URL.String())
	})

	c.OnHTML(`div.product_listing_box`, func(e *colly.HTMLElement) {
		origin := e.DOM.Parent().Children().First().Find("a:nth-child(3)").Text()
		name := e.ChildText("a.product_listing_box_name")
		url := e.ChildAttr("a.product_listing_box_name", "href")

		log.Println("Scraping ", name)

		cigar := Cigar{
			site:   "hacico",
			origin: origin,
			name:   name,
			url:    url,
		}

		e.ForEach("table table tr", func(i int, el *colly.HTMLElement) {
			availability := getCigarAvailability(el)

			if i == 0 {
				cigar.single = availability
			} else {
				cigar.box = availability
			}
		})

		cigars = append(cigars, cigar)
	})

	c.Visit("https://www.hacico.de/en/Cigars/Nicaragua")
	c.Wait()

	return cigars
}
