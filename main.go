package main

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/caarlos0/env/v6"
	"github.com/gocolly/colly"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// 	dsn := "root:@tcp(127.0.0.1:3306)/rozetka?charset=utf8mb4&parseTime=True&loc=Local"
//         {Home:/your/home Port:3000 IsProduction:true Hosts:[host1 host2 host3] Duration:1s}
type config struct {
	Home         string        `env:"HOME"`
	Port         int           `env:"PORT" envDefault:"3306"`
	Password     string        `env:"unset"`
	IsProduction bool          `env:"PRODUCTION"`
	Hosts        []string      `env:"HOSTS" envSeparator:":"`
	Duration     time.Duration `env:"DURATION"`
	TempFolder   string        `env:"TEMP_FOLDER" envDefault:"${HOME}/tmp" envExpand:"true"`
}

type Product struct {
	gorm.Model
	ID         int
	OriginalID string
	Name       string
}

type ProductChar struct {
	gorm.Model
	ProductID int
	Name      string
	Value     string
}

// todo move dsn variable to env and pictures to env
// https://github.com/caarlos0/env
func main() {
	// ENV Carloos
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}

	fmt.Printf("CONFIGGGG  %+v\n", cfg)

	// смотрите https://github.com/go-sql-driver/mysql#dsn-data-source-name для подробностей
	//dsn := "root:@tcp(127.0.0.1:3306)/rozetka?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(cfg.Home), &gorm.Config{})
	if err != nil {
		return
	}

	// Миграция схем
	if err = db.AutoMigrate(&Product{}, &ProductChar{}); err != nil {
		return
	}

	//check folder exists pictures
	if _, err := os.Stat("pictures"); err != nil {
		if os.IsNotExist(err) {
			err := os.Mkdir("pictures", 0700)
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

		product := Product{
			Name:       name,
			OriginalID: code,
		}

		db.Create(&product)

		//check pictures folder exist, if not created, create
		path := "pictures/" + code
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

			db.Create(&ProductChar{
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

	err = c.Visit("https://rozetka.com.ua/all-tv/c80037/sell_status=available;seller=rozetka/")
	if err != nil {
		return
	}
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

// save all picures into this folder
// add id from rozetka to products table (Код:  293091343) - original_id
// after save of the new product if you will have any pictures,
// put all pictures
