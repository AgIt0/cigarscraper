package main

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
