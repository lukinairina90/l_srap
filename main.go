package main

import (
	"fmt"
	"github.com/gocolly/colly"
)

func main() {
	c := colly.NewCollector()
	// Find and visit all links
	c.OnHTML(".goods-tile__picture", func(e *colly.HTMLElement) {
		err := e.Request.Visit(e.Attr("href"))
		if err != nil {
			return
		}
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	//c.OnHTML(`a[href]`, func(e *colly.HTMLElement) {
	//	e.Request.Visit(e.Attr(".tabs__link"))
	//})

	// Find and visit next page links
	c.OnHTML(`a:contains(' Характеристики ')`, func(e *colly.HTMLElement) {
		e
	})

	c.OnRequest(func(e *colly.Request) {
		fmt.Println("Charact", e.URL)
	})

	err := c.Visit("https://rozetka.com.ua/all-tv/c80037/sell_status=available;seller=rozetka/")
	if err != nil {
		return
	}
}
