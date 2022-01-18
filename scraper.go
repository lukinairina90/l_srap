package main

import (
	"R_Scraper/models"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type scraper struct {
	cfg Config
	db  *gorm.DB
}

func (s *scraper) collectData() error {
	//check folder exists pictures
	if _, err := os.Stat(s.cfg.PictureFolder); err != nil {
		if os.IsNotExist(err) {
			err := os.Mkdir(s.cfg.PictureFolder, 0700)
			if err != nil {
				log.Println(err)
			}
		}
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

		codeEl := e.DOM.Closest("body").Find(".product__code")
		codeEl.Find("span").Remove()
		code := strings.TrimSpace(codeEl.Text())

		product := models.Product{
			Name:       name,
			OriginalID: code,
		}

		s.db.Create(&product)

		//check pictures folder exist, if not created, create
		path := s.cfg.PictureFolder + "/" + code
		err := os.Mkdir(path, 0700)
		if err != nil {
			log.Println(err)
		}

		e.DOM.Closest("body").Find(".thumbnail__picture[src*='images']").Each(func(i int, s *goquery.Selection) {
			thumbSrc, _ := s.Attr("src")
			thumbSplit := strings.Split(thumbSrc, "/")
			thumbPic := thumbSplit[len(thumbSplit)-1]
			thumbPic = thumbPic[:len(thumbPic)-4]

			err := downloadFile(thumbSrc, path, thumbPic)
			if err != nil {
				log.Fatal(err)
			}
		})

		e.ForEach("body.thumbnail__picture", func(_ int, e *colly.HTMLElement) {
			elName := e.DOM.Find(".a > img").Text()
			println(elName)
		})

		//Iterate over rows of the table which contains different information
		//about the course
		e.ForEach(".characteristics-full__item", func(_ int, el *colly.HTMLElement) {
			charName := el.DOM.Find(".characteristics-full__label span").Text()
			charValue := el.DOM.Find("a.ng-star-inserted").Text()

			s.db.Create(&models.ProductChar{
				ProductID: product.ID,
				Name:      charName,
				Value:     charValue,
			})
		})
	})

	//pagination
	c.OnHTML(`.pagination__direction--forward`, func(e *colly.HTMLElement) {
		println("Next page link found:", e.Attr("href"))
		err := e.Request.Visit(e.Attr("href"))
		if err != nil {
			return
		}
	})

	return c.Visit("https://rozetka.com.ua/all-tv/c80037/sell_status=available;seller=rozetka/")
}

func downloadFile(url, path, picID string) error {
	//Get the response bytes from the url
	//todo get full size pictures
	//https://content.rozetka.com.ua/goods/images/big/237518862.jpg
	//https://content.rozetka.com.ua/goods/images/preview/240240612.jpg
	url = strings.Replace(url, "preview/", "big/", 1)
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New("received non 200 response code")
	}
	//Create an empty file
	file, err := os.Create(path + "/" + picID + ".jpg")
	if err != nil {
		return err
	}
	defer file.Close()

	//Write the bytes to the fiel
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}
