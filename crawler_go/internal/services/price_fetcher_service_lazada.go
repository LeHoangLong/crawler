package services

import (
	"context"
	"crawler_go/internal/controllers"
	"crawler_go/internal/models"
	"crawler_go/internal/repositories"
	"crawler_go/internal/views"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

type PriceFetcherServiceLazada struct {
	pageRepository *repositories.PageRepository
}

func MakePriceFetcherServiceLazada(
	iPageRepository *repositories.PageRepository,
) PriceFetcherServiceLazada {
	return PriceFetcherServiceLazada{
		pageRepository: iPageRepository,
	}
}

func (s *PriceFetcherServiceLazada) fetchPrice(ctx context.Context, keyword string, iPage *playwright.Page) ([]models.ItemPrice, error) {
	page := *iPage
	waitUntil := "domcontentloaded"
	option := playwright.PageGotoOptions{
		WaitUntil: ((*playwright.WaitUntilState)(&waitUntil)),
	}

	var err error
	pageLoaded := false
	for !pageLoaded {
	start:
		if _, err = page.Goto("https://www.lazada.vn/", option); err != nil {
			return []models.ItemPrice{}, err
		}

		searchBoxFound := false
		var textInput *playwright.ElementHandle
		start := time.Now()
		for !searchBoxFound {
			entries, err := page.QuerySelectorAll("input")

			if err == nil {
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

			now := time.Now()
			if now.Sub(start) > 5*time.Second {
				goto start /// restart
			}
		}
		fmt.Println(2)

		searchButtonFound := false
		var searchButton *playwright.ElementHandle
		start = time.Now()
		for !searchButtonFound {
			entries, err := page.QuerySelectorAll("button")
			if err == nil {
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

			now := time.Now()
			if now.Sub(start) > 5*time.Second {
				goto start /// restart
			}
		}

		fmt.Println(3)

		(*textInput).Type(keyword)
		err = (*searchButton).Click()
		if err != nil {
			log.Fatalf("could not click: %v", err)
		}

		dataFound := false
		start = time.Now()
		for !dataFound {
			entries, err := page.QuerySelectorAll("div")
			if err == nil {
				for i := range entries {
					attr, err := entries[i].GetAttribute("data-qa-locator")
					if err == nil {
						if attr == "product-item" {
							dataFound = true
							break
						}
					}
				}
			}

			now := time.Now()
			if now.Sub(start) > 5*time.Second {
				goto start /// restart
			}
			time.Sleep(1 * time.Second)
		}

		fmt.Println(4)
		if dataFound {
			pageLoaded = true
		}
	}
	fmt.Println(5)

	items := []models.ItemPrice{}
	{
		entries, err := page.QuerySelectorAll("div")
		if err != nil {
			return []models.ItemPrice{}, err
		}

		for i := range entries {
			if len(items) > 8 {
				break
			}
			attr, err := entries[i].GetAttribute("data-qa-locator")
			if err == nil {
				if attr == "product-item" {
					minPrice := ""
					maxHeight := 0
					maxHeightIndex := -1
					/// extract item

					/// get price
					spans, err := entries[i].QuerySelectorAll("span")
					if err != nil {
						return []models.ItemPrice{}, err
					}

					for i := range spans {
						text, err := spans[i].TextContent()
						if err != nil {
							return []models.ItemPrice{}, err
						}

						if strings.Contains(text, "₫") {
							box, err := spans[i].BoundingBox()
							if err != nil {
								return []models.ItemPrice{}, err
							}
							height := box.Height
							if height > maxHeight {
								maxHeightIndex = i
							}
						}
					}

					if maxHeightIndex != -1 {
						minPrice, _ = spans[maxHeightIndex].TextContent()
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

					newItem := models.MakeItemPrice(title, minPrice, "", link, image)
					items = append(items, newItem)
				}
			}
		}
	}

	return items, nil
}

func (s *PriceFetcherServiceLazada) FetchPrice(ctx context.Context, option views.SearchOption, results chan controllers.ItemPriceResult) {
	page, err := s.pageRepository.GetPage(ctx)
	defer s.pageRepository.ReturnPage(page)
	if err != nil {
		results <- controllers.ItemPriceResult{
			Prices: []models.ItemPrice{},
			Err:    err,
		}
		return
	}

	fmt.Println("fetching lazada")
	start := time.Now()
	testPrices, err := s.fetchPrice(ctx, option.Keyword, page)
	end := time.Now()
	fmt.Println("lazada fetched, ", end.Sub(start))
	/*
		testPrices := []models.ItemPrice{
			models.MakeItemPrice(
				"Bamboo Shoe Rack Bench/Seat Wearing Taking off Shoes Strong Organizer",
				"$19.80",
				"$36.80",
				"https://shopee.sg/Bamboo-Shoe-Rack-Bench-Seat-Wearing-Taking…8fc8-c5c0d58c9776&xptdk=d9b58e72-7371-48d0-8fc8-c5c0d58c9776",
				"https://cf.shopee.sg/file/b60552d9a6e4f2291d9fea21bbd1b0e0_tn",
			),
			models.MakeItemPrice(
				"Bamboo Shoe Rack Bench/Seat Wearing Taking off Shoes Strong Organizer",
				"$19.80",
				"",
				"https://shopee.sg/Bamboo-Shoe-Rack-Bench-Seat-Wearing-Taking…8fc8-c5c0d58c9776&xptdk=d9b58e72-7371-48d0-8fc8-c5c0d58c9776",
				"https://cf.shopee.sg/file/b60552d9a6e4f2291d9fea21bbd1b0e0_tn",
			),
		}
	*/

	if err != nil {
		results <- controllers.ItemPriceResult{
			Prices: []models.ItemPrice{},
			Err:    err,
		}
		return
	}
	results <- controllers.ItemPriceResult{
		Prices: testPrices,
		Err:    nil,
	}
}
