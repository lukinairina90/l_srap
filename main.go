package main

import (
	"fmt"
	"github.com/gocolly/colly"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
	"strings"
)

type Product struct {
	gorm.Model
	ID   int
	Name string
}

type ProductChar struct {
	gorm.Model
	ProductID int
	Name      string
	Value     string
}

func main() {
	// смотрите https://github.com/go-sql-driver/mysql#dsn-data-source-name для подробностей
	dsn := "root:@tcp(127.0.0.1:3306)/rozetka?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return
	}

	// Миграция схем
	if err = db.AutoMigrate(&Product{}, &ProductChar{}); err != nil {
		return
	}

	c := colly.NewCollector()
	//Find and visit all links
	c.OnHTML(".goods-tile__picture", func(e *colly.HTMLElement) {
		err := e.Request.Visit(e.Attr("href"))
		if err != nil {
			return
		}
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	//Find and visit next page links
	c.OnHTML(`.product-about__characteristics`, func(e *colly.HTMLElement) {
		name := e.DOM.Find(".product-tabs__heading_color_gray").Text()
		//todo write into products
		product := Product{
			Name: name,
		}

		db.Create(&product)

		codeEl := e.DOM.Closest("body").Find(".product__code")
		codeEl.Find("span").Remove()
		code := strings.TrimSpace(codeEl.Text())
		//check picture folder exist, if not created, create

		path := "pictures/" + code
		err := os.Mkdir(path, 0700)
		if err != nil {
			log.Println(err)
		}

		//https://content.rozetka.com.ua/goods/images/big/237518862.jpg
		e.DOM.Closest("body").Find("картинка").Each()

		// Iterate over rows of the table which contains different information
		// about the course
		//e.ForEach(".characteristics-full__item", func(_ int, el *colly.HTMLElement) {
		//	charName := el.DOM.Find(".characteristics-full__label span").Text()
		//	charValue := el.DOM.Find("a.ng-star-inserted").Text()
		//
		//	db.Create(&ProductChar{
		//		ProductID: product.ID,
		//		Name:      charName,
		//		Value:     charValue,
		//	})
		//})
	})

	//pagination
	//c.OnHTML(`.pagination__direction--forward`, func(e *colly.HTMLElement) {
	//	println("Next page link found:", e.Attr("href"))
	//	err := e.Request.Visit(e.Attr("href"))
	//	if err != nil {
	//		return
	//	}
	//})

	err = c.Visit("https://rozetka.com.ua/all-tv/c80037/sell_status=available;seller=rozetka/")
	if err != nil {
		return
	}
}

// todo add pictures folder(don`t forget about gitignore)
// save all picures into this folder
// add id from rozetka to products table (Код:  293091343) - original_id
// after save of the new product if you will have any pictures,
// create folder named as id from rozetka
// put all pictures
