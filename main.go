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
	c.OnHTML(`.characteristics-full__list`, func(e *colly.HTMLElement) {
		// Iterate over rows of the table which contains different information
		// about the course
		e.ForEach(".characteristics-full__item", func(_ int, el *colly.HTMLElement) {
			label := el.DOM.Find(".characteristics-full__label span").Text()
			char := el.DOM.Find("a.ng-star-inserted").Text()

			println(label + ": " + char)
		})
	})

	//todo add pagination
	//add c.OnHTML to check pagination .pagination__list
	//find next arrow item, check if it doesn`t have attr disabled or just check if href is present
	//if not disabled, just take href from arrow and visit it
	err := c.Visit("https://rozetka.com.ua/all-tv/c80037/sell_status=available;seller=rozetka/")
	if err != nil {
		return
	}
}
