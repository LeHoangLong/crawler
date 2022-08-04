package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

type ItemPrice struct {
	Title string
	Max   string
	Min   string
	Link  string
	Image string
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

	waitUntil := "domcontentloaded"
	option := playwright.PageGotoOptions{
		WaitUntil: ((*playwright.WaitUntilState)(&waitUntil)),
	}
	if _, err = page.Goto("https://www.lazada.vn/", option); err != nil {
		log.Fatalf("could not goto: %v", err)
	}

	time.Sleep(1000 * time.Millisecond)

	// os.WriteFile("test.txt", []byte(text), 0666)
	fmt.Println(1)
	searchBoxFound := false
	var textInput *playwright.ElementHandle
	for !searchBoxFound {
		entries, err := page.QuerySelectorAll("input")
		if err != nil {
			log.Fatalf("could not get entries: %v", err)
		}

		for i := range entries {
			attr, err := entries[i].GetAttribute("type")
			if err == nil {
				if attr == "search" {
					searchBoxFound = true
					textInput = &entries[i]
				}
			}
		}
	}
	fmt.Println(2)

	searchButtonFound := false
	var searchButton *playwright.ElementHandle
	for !searchButtonFound {
		entries, err := page.QuerySelectorAll("button")
		if err != nil {
			log.Fatalf("could not get entries: %v", err)
		}

		for i := range entries {
			attr, err := entries[i].GetAttribute("class")
			if err == nil {
				if strings.Contains(attr, "search-box") {
					searchButtonFound = true
					searchButton = &entries[i]
				}
			}
		}
	}
	fmt.Println(3)

	(*textInput).Type("áo")
	err = (*searchButton).Click()
	/// err = page.Keyboard().Press("Enter")
	if err != nil {
		log.Fatalf("could not click: %v", err)
	}

	dataFound := false
	for !dataFound {
		entries, err := page.QuerySelectorAll("div")
		if err != nil {
			log.Fatalf("could not get div: %v", err)
		}

		for i := range entries {
			attr, err := entries[i].GetAttribute("data-qa-locator")
			if err == nil {
				if attr == "product-item" {
					dataFound = true
					break
				}
			}
		}
		time.Sleep(1 * time.Second)
	}
	fmt.Println(4)

	image, err := page.Screenshot()
	if err != nil {
		log.Fatalf("could not screenshot: %v", err)
	}

	os.WriteFile("test.png", image, 0666)

	items := []ItemPrice{}
	{
		entries, err := page.QuerySelectorAll("div")
		if err != nil {
			log.Fatalf("could not get div: %v", err)
		}

		for i := range entries {
			attr, err := entries[i].GetAttribute("data-qa-locator")
			if err == nil {
				if attr == "product-item" {
					newItem := ItemPrice{}
					maxHeight := 0
					maxHeightIndex := -1
					/// extract item
					spans, err := entries[i].QuerySelectorAll("span")
					if err != nil {
						log.Fatalf("could not get spans: %v", err)
					}

					for i := range spans {
						text, err := spans[i].TextContent()
						if err != nil {
							log.Fatalf("could not get text: %v", err)
						}

						if strings.Contains(text, "₫") {
							box, err := spans[i].BoundingBox()
							if err != nil {
								log.Fatalf("could not get box: %v", err)
							}
							height := box.Height
							if height > maxHeight {
								maxHeightIndex = i
							}
						}
					}

					/// get title
					title := ""
					link := ""
					{
						maxLength := 0
						maxLengthIndex := -1
						links, err := entries[i].QuerySelectorAll("a")
						if err != nil {
							log.Fatalf("could not get link: %v", err)
						}

						for i := range links {
							text, err := links[i].TextContent()
							if err != nil {
								log.Fatalf("could not get text: %v", err)
							}

							if len(text) > maxLength {
								maxLength = len(text)
								maxLengthIndex = i
							}
						}

						if maxLengthIndex != -1 {
							title, _ = links[maxLengthIndex].TextContent()
							link, _ = links[maxLengthIndex].GetAttribute("href")
							link = "https:" + link
						}
					}

					/// get image
					image := ""
					{
						maxHeight := 0
						maxHeightIndex := -1
						imgs, err := entries[i].QuerySelectorAll("img")
						if err != nil {
							log.Fatalf("could not get img: %v", err)
						}

						for i := range imgs {
							box, err := imgs[i].BoundingBox()
							if err != nil {
								log.Fatalf("could not get bb: %v", err)
							}

							height := box.Height
							if height > maxHeight {
								maxHeight = height
								maxHeightIndex = i
							}
						}

						if maxHeightIndex != -1 {
							image, _ = imgs[maxHeightIndex].GetAttribute("src")
						}
					}

					if maxHeightIndex != -1 {
						text, _ := spans[maxHeightIndex].TextContent()
						newItem.Min = text
						newItem.Title = title
						newItem.Link = link
						newItem.Image = image
						items = append(items, newItem)
					}
				}
			}
		}
	}

	fmt.Printf("items: %+v", items)
	/*
		entries, err := page.QuerySelectorAll("input")
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
	*/
}
