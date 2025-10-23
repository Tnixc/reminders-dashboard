package main

import (
	"fmt"
	"math/rand"
)

// randomItemGenerator generates random items for the list
type randomItemGenerator struct {
	titles       []string
	descriptions []string
	titleIndex   int
	descIndex    int
}

func (r *randomItemGenerator) reset() {
	r.titles = []string{
		"Ramen",
		"Tomato Soup",
		"Hamburgers",
		"Cheeseburgers",
		"Currywurst",
		"Fish and Chips",
		"Łazanki",
		"Lobster",
		"Pasta",
		"Pizza",
		"Noodles",
		"Sushi",
		"Tacos",
		"Dumplings",
		"Burritos",
		"Pad Thai",
		"Biryani",
		"Pho",
		"Falafel",
		"Kebab",
		"Dim Sum",
		"Ratatouille",
		"Goulash",
		"Paella",
	}

	r.descriptions = []string{
		"Yummy",
		"Delicious",
		"Tasty",
		"Mouth-watering",
		"Scrumptious",
		"Delectable",
		"Savory",
		"Flavorful",
	}

	r.titleIndex = 0
	r.descIndex = 0
}

func (r *randomItemGenerator) next() item {
	if r.titles == nil || len(r.titles) == 0 {
		r.reset()
	}

	if r.titleIndex >= len(r.titles) {
		r.titleIndex = 0
	}

	if r.descIndex >= len(r.descriptions) {
		r.descIndex = 0
	}

	title := r.titles[r.titleIndex]
	desc := r.descriptions[r.descIndex]

	r.titleIndex++
	r.descIndex = rand.Intn(len(r.descriptions))

	return item{
		title:       title,
		description: fmt.Sprintf("%s %s", desc, title),
	}
}
