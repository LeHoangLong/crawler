package main_sample_shopee

import (
	"fmt"
	"log"
	"time"

	"github.com/playwright-community/playwright-go"
)

type ItemPrice struct {
	Title string
	Max   string
	Min   string
	Link  string
}

func main() {
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	browser, err := pw.Chromium.Launch()
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}
	page, err := browser.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}
	if _, err = page.Goto("https://shopee.sg/search?keyword=shoe"); err != nil {
		log.Fatalf("could not goto: %v", err)
	}

	time.Sleep(1000 * time.Millisecond)

	entries, err := page.QuerySelectorAll(".shopee-search-item-result__item")
	if err != nil {
		log.Fatalf("could not get entries: %v", err)
	}

	fmt.Println(len(entries))

	prices := []ItemPrice{}
	for i, entry := range entries {
		if i > 5 {
			break
		}
		entry.ScrollIntoViewIfNeeded()
		entry.Focus()
		spans, err := entry.QuerySelectorAll("span")
		if err != nil {
			log.Fatalf("could not get entries: %v", err)
		}

		itemPrice := ItemPrice{
			Max: "",
			Min: "",
		}
		signFound := false
		for _, span := range spans {
			text, err := span.TextContent()
			if err != nil {
				log.Fatalf("could not get text: %v", err)
			}

			if text == "$" {
				signFound = true
			} else if signFound {
				signFound = false
				if itemPrice.Min == "" {
					itemPrice.Min = text
				} else {
					itemPrice.Max = text
				}
			}
		}

		/// find image bottom y
		img, err := entry.QuerySelector("img")
		if err != nil {
			log.Fatalf("could not get img: %v", err)
		}
		rect, err := img.BoundingBox()
		if err != nil {
			log.Fatalf("could not get rect: %v", err)
		}
		imgBottomY := rect.Y + rect.Height

		/// find text closest to image
		divs, err := entry.QuerySelectorAll("div")
		highestY := 10000000
		highestYIndex := -1
		title := ""
		for i := range divs {
			text, err := divs[i].TextContent()
			if err != nil {
				log.Fatalf("could not get text: %v", err)
			}

			if len(text) == 0 {
				continue
			}

			rect, err := divs[i].BoundingBox()
			if err != nil {
				log.Fatalf("could not get rect: %v", err)
			}
			bottomY := rect.Y + rect.Height
			if bottomY <= imgBottomY {
				continue
			}

			if bottomY <= highestY || highestYIndex == -1 {
				highestYIndex = i
				highestY = bottomY
				title = text
			}
		}

		itemPrice.Title = title

		/// find link
		{
			a, err := entry.QuerySelector("a")
			if err != nil {
				log.Fatalf("could not get a: %v", err)
			}
			link, err := a.GetAttribute("href")
			if err != nil {
				log.Fatalf("could not get href: %v", err)
			}
			itemPrice.Link = link
		}

		prices = append(prices, itemPrice)
	}
	fmt.Printf("prices: %+v", prices)
	fmt.Println("len: ", len(prices))

	if err = browser.Close(); err != nil {
		log.Fatalf("could not close browser: %v", err)
	}
	if err = pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}
	fmt.Println("hello")
}
