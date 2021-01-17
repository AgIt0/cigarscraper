package main

import (
	"encoding/csv"
	"log"
	"os"
	"strconv"
)

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
		writer.Write(
			[]string{
				cigar.site,
				cigar.origin,
				cigar.name,
				cigar.url,
				strconv.FormatFloat(cigar.single.price, 'g', -1, 64),
				cigar.single.currency,
				strconv.FormatBool(cigar.single.available),
				strconv.FormatFloat(cigar.box.price, 'g', -1, 64),
				cigar.box.currency,
				strconv.FormatBool(cigar.box.available),
			},
		)
	}
}

func main() {
	cigars := scrapeBestCigars()
	cigars = append(cigars, scrapeHacico()...)
	log.Println(cigars)
	saveToCSV(cigars)
}
