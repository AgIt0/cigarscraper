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
	site   string
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

	writer.Write([]string{"site", "origin", "name", "url", "single_price", "single_currency", "single_available", "box_price", "box_currency", "box_available"})
	for _, cigar := range cigars {
		writer.Write([]string{cigar.site, cigar.origin, cigar.name, cigar.url, strconv.FormatFloat(cigar.single.price, 'g', -1, 64), cigar.single.currency, strconv.FormatBool(cigar.single.available), strconv.FormatFloat(cigar.box.price, 'g', -1, 64), cigar.box.currency, strconv.FormatBool(cigar.box.available)})
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
func main() {
	cigars := scrapeBestCigars()
	cigars = append(cigars, scrapeHacico()...)
	log.Println(cigars)
	saveToCSV(cigars)
}
