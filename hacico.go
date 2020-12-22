package main

import (
	"encoding/csv"
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
	origin string
	name   string
	url    string
	single Availability
	box    Availability
}

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

func saveToCSV(cigars []Cigar) {
	// just delete/recreate if existing
	os.Remove("cigars.csv")
	file, err := os.Create("cigars.csv")
	defer file.Close()
	if err != nil {
		log.Fatalln("failed to open file", err)
	}

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"origin", "name", "url", "single_price", "single_currency", "single_available", "box_price", "box_currency", "box_available"})
	for _, cigar := range cigars {
		writer.Write([]string{cigar.origin, cigar.name, cigar.url, strconv.FormatFloat(cigar.single.price, 'g', -1, 64), cigar.single.currency, strconv.FormatBool(cigar.single.available), strconv.FormatFloat(cigar.box.price, 'g', -1, 64), cigar.box.currency, strconv.FormatBool(cigar.box.available)})
	}
}

func main() {
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
		origin := e.DOM.Parent().Children().First().Find("a:nth-child(3)").Text()
		name := e.ChildText("a.product_listing_box_name")
		url := e.ChildAttr("a.product_listing_box_name", "href")

		log.Println("Scraping ", name)

		cigar := Cigar{
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

		log.Println(cigar)
		cigars = append(cigars, cigar)
	})

	// start scraping
	c.Visit("https://www.hacico.de/en/Cigars/Nicaragua")

	saveToCSV(cigars)
}
