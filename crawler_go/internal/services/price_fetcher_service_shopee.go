package services

import (
	"context"
	"crawler_go/internal/controllers"
	"crawler_go/internal/models"
	"crawler_go/internal/repositories"
	"crawler_go/internal/views"
	"fmt"
	"time"

	"github.com/playwright-community/playwright-go"
)

type PriceFetcherServiceShopee struct {
	pageRepository *repositories.PageRepository
}

func MakePriceFetcherServiceShopee(
	iPageRepository *repositories.PageRepository,
) PriceFetcherServiceShopee {
	return PriceFetcherServiceShopee{
		pageRepository: iPageRepository,
	}
}

func (s *PriceFetcherServiceShopee) fetchPrice(ctx context.Context, keyword string, iPage *playwright.Page) ([]models.ItemPrice, error) {
	if _, err := (*iPage).Goto("https://shopee.vn/search?keyword=" + keyword); err != nil {
		return []models.ItemPrice{}, err
	}

	var err error
	entries := []playwright.ElementHandle{}
	for len(entries) == 0 {
		entries, err = (*iPage).QuerySelectorAll(".shopee-search-item-result__item")
		if err != nil {
			return []models.ItemPrice{}, err
		}

		if len(entries) == 0 {
			time.Sleep(time.Millisecond * 500)
		}
	}

	prices := []models.ItemPrice{}
	for _, entry := range entries {
		if len(prices) > 8 {
			break
		}

		entry.ScrollIntoViewIfNeeded()
		entry.Focus()
		spans, err := entry.QuerySelectorAll("span")
		if err != nil {
			return []models.ItemPrice{}, err
		}

		min := ""
		max := ""
		url := ""
		image := ""
		signFound := false
		for _, span := range spans {
			text, err := span.TextContent()
			if err != nil {
				return []models.ItemPrice{}, err
			}

			if text == "₫" {
				signFound = true
			} else if signFound {
				signFound = false
				if min == "" {
					min = text + " ₫"
				} else {
					max = text + " ₫"
				}
			}
		}

		/// find image bottom y
		img, err := entry.QuerySelector("img")
		if err != nil {
			return []models.ItemPrice{}, err
		}
		rect, err := img.BoundingBox()
		if err != nil {
			return []models.ItemPrice{}, err
		}
		image, err = img.GetAttribute("src")
		if err != nil {
			return []models.ItemPrice{}, err
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
				return []models.ItemPrice{}, err
			}

			if len(text) == 0 {
				continue
			}

			rect, err := divs[i].BoundingBox()
			if err != nil {
				return []models.ItemPrice{}, err
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

		{
			a, err := entry.QuerySelector("a")
			if err != nil {
				return []models.ItemPrice{}, err
			}
			link, err := a.GetAttribute("href")
			if err != nil {
				return []models.ItemPrice{}, err
			}
			url = "https://shopee.vn" + link
		}

		itemPrice := models.MakeItemPrice(title, min, max, url, image)

		prices = append(prices, itemPrice)
	}

	return prices, nil
}

func (s *PriceFetcherServiceShopee) FetchPrice(ctx context.Context, option views.SearchOption, results chan controllers.ItemPriceResult) {
	page, err := s.pageRepository.GetPage(ctx)
	defer s.pageRepository.ReturnPage(page)
	if err != nil {
		results <- controllers.ItemPriceResult{
			Prices: []models.ItemPrice{},
			Err:    err,
		}
		return
	}

	fmt.Println("fetching shopee")
	start := time.Now()
	testPrices, err := s.fetchPrice(ctx, option.Keyword, page)
	end := time.Now()
	fmt.Println("shopee fetched, ", end.Sub(start))
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
