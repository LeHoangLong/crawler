package services

import (
	"context"
	"crawler_go/internal/controllers"
	"crawler_go/internal/models"
	"crawler_go/internal/repositories"
	"crawler_go/internal/views"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/playwright-community/playwright-go"

	"net/url"
)

type PriceFetcherServiceLazada struct {
	pageRepository      *repositories.PageRepository
	additionalUrlParams string
	mutex               *sync.Mutex
}

func MakePriceFetcherServiceLazada(
	iPageRepository *repositories.PageRepository,
) PriceFetcherServiceLazada {
	service := PriceFetcherServiceLazada{
		pageRepository:      iPageRepository,
		additionalUrlParams: "",
		mutex:               &sync.Mutex{},
	}
	service.init()
	return service
}

func (s *PriceFetcherServiceLazada) init() {
	ctx := context.Background()
	page, err := s.pageRepository.GetPage(ctx)
	if err != nil {
		return
	}
	defer s.pageRepository.ReturnPage(page)
	s.getCatalogPage("áo", page)
}

func (s *PriceFetcherServiceLazada) getCatalogPage(iKeyword string, iPage *playwright.Page) error {
	waitUntil := "domcontentloaded"
	option := playwright.PageGotoOptions{
		WaitUntil: ((*playwright.WaitUntilState)(&waitUntil)),
	}

	var err error
	pageLoaded := false
	beforeFetchPrice := time.Now()
	s.mutex.Lock()
	additionalUrlParams := s.additionalUrlParams
	s.mutex.Unlock()
	for !pageLoaded {
	start:
		{
			now := time.Now()
			if now.Sub(beforeFetchPrice).Seconds() > 60 {
				additionalUrlParams = ""
				return fmt.Errorf("could not load page from lazada")
			}
		}

		if additionalUrlParams != "" {
			catalogUrl := "https://www.lazada.vn/catalog/?q="
			catalogUrl += url.QueryEscape(iKeyword)
			catalogUrl += "&" + s.additionalUrlParams
			_, err := (*iPage).Goto(catalogUrl, option)
			if err != nil {
				return err
			}
		} else {
			if _, err = (*iPage).Goto("https://www.lazada.vn/", option); err != nil {
				return err
			}

			searchBoxFound := false
			var textInput *playwright.ElementHandle
			start := time.Now()
			for !searchBoxFound {
				entries, err := (*iPage).QuerySelectorAll("input")

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
				if now.Sub(start) > 15*time.Second {
					goto start /// restart
				}
			}

			searchButtonFound := false
			var searchButton *playwright.ElementHandle
			start = time.Now()
			for !searchButtonFound {
				entries, err := (*iPage).QuerySelectorAll("button")
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
				if now.Sub(start) > 15*time.Second {
					goto start /// restart
				}
			}

			fmt.Println(3)

			(*textInput).Type(iKeyword)
			(*searchButton).Click()
		}

		dataFound := false
		start := time.Now()
		for !dataFound {
			entries, err := (*iPage).QuerySelectorAll("div")
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
			if now.Sub(start) > 15*time.Second {
				goto start /// restart
			}
			time.Sleep(1 * time.Second)
		}

		fmt.Println(4)
		if dataFound {
			pageLoaded = true
		}
	}

	if pageLoaded {
		rawUrl := (*iPage).URL()
		parsedUrl, err := url.Parse(rawUrl)
		if err != nil {
			return nil
		}

		query, _ := url.ParseQuery(parsedUrl.RawQuery)
		query.Del("q")
		s.mutex.Lock()
		s.additionalUrlParams = query.Encode()
		s.mutex.Unlock()
	}

	return nil
}

func (s *PriceFetcherServiceLazada) fetchPrice(ctx context.Context, keyword string, iPage *playwright.Page) ([]models.ItemPrice, error) {
	err := s.getCatalogPage(keyword, iPage)
	if err != nil {
		return []models.ItemPrice{}, nil
	}

	items := []models.ItemPrice{}
	{
		entries, err := (*iPage).QuerySelectorAll("div")
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
					entries[i].ScrollIntoViewIfNeeded()
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
							return []models.ItemPrice{}, err
						}

						for i := range links {
							text, err := links[i].TextContent()
							if err != nil {
								return []models.ItemPrice{}, err
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
					imageRendered := false
					image := ""
					for !imageRendered {
						entries[i].ScrollIntoViewIfNeeded()
						{
							imageIndex := -1
							imgs, err := entries[i].QuerySelectorAll("img")
							if err != nil {
								return []models.ItemPrice{}, err
							}

							for i := range imgs {
								typeVal, err := imgs[i].GetAttribute("type")
								if err != nil {
									return []models.ItemPrice{}, err
								}

								if typeVal == "product" {
									imageIndex = i
									break
								}
							}

							if imageIndex != -1 {
								image, _ = imgs[imageIndex].GetAttribute("src")
							}

							if image != "" && !strings.Contains(image, "data:image/png;") {
								imageRendered = true
							} else {
								time.Sleep(500 * time.Millisecond)
							}
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
	if err != nil {
		results <- controllers.ItemPriceResult{
			Prices: []models.ItemPrice{},
			Err:    err,
		}
		return
	}
	defer s.pageRepository.ReturnPage(page)

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
