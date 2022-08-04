package controllers

import (
	"context"
	"crawler_go/internal/models"
	"crawler_go/internal/views"
	"fmt"
	"sync"
)

type ItemPriceResult struct {
	Err    error
	Prices []models.ItemPrice
}

type PriceFetcherServiceI interface {
	FetchPrice(ctx context.Context, option views.SearchOption, results chan ItemPriceResult)
}

type priceRequest struct {
	id     string
	option views.SearchOption
	ctx    context.Context
}

type PriceRequestController struct {
	requestQueue    []priceRequest
	isStopped       bool
	mutex           *sync.Mutex
	cond            *sync.Cond
	fetchers        []PriceFetcherServiceI
	completeHandler views.PriceRequestCompleteHandler
	requestCounter  uint32
}

func MakePriceRequestController() PriceRequestController {
	mutex := sync.Mutex{}
	return PriceRequestController{
		requestQueue:    []priceRequest{},
		isStopped:       false,
		mutex:           &mutex,
		cond:            sync.NewCond(&mutex),
		fetchers:        []PriceFetcherServiceI{},
		completeHandler: nil,
		requestCounter:  0,
	}
}

func (c *PriceRequestController) AddPriceFetcherService(
	iService PriceFetcherServiceI,
) {
	c.mutex.Lock()
	c.fetchers = append(c.fetchers, iService)
	c.mutex.Unlock()
}

func (c *PriceRequestController) SubmitPriceRequest(ctx context.Context, option views.SearchOption) models.PriceRequestId {
	fmt.Println("start submit price request")
	c.mutex.Lock()
	counter := c.requestCounter
	c.requestCounter++
	id := fmt.Sprintf("%d", counter)
	c.mutex.Unlock()
	newRequest := priceRequest{
		id:     id,
		option: option,
		ctx:    ctx,
	}
	c.mutex.Lock()
	c.requestQueue = append(c.requestQueue, newRequest)
	c.cond.Broadcast()
	c.mutex.Unlock()
	return id
}

func (c *PriceRequestController) SetCompleteHandler(handler views.PriceRequestCompleteHandler) {
	c.mutex.Lock()
	c.completeHandler = handler
	c.mutex.Unlock()
}

func (c *PriceRequestController) Start() {
	c.mutex.Lock()
	c.isStopped = false
	go c.process()
	c.mutex.Unlock()
}

func (c *PriceRequestController) Stop() {
	c.mutex.Lock()
	c.isStopped = true
	c.cond.Broadcast()
	c.mutex.Unlock()
}

func (c *PriceRequestController) process() {
	for !c.isStopped {
		c.mutex.Lock()
		for len(c.requestQueue) == 0 && !c.isStopped {
			c.cond.Wait()
		}
		requestQueue := c.requestQueue
		c.requestQueue = []priceRequest{}
		c.mutex.Unlock()

		if c.isStopped {
			return
		}

		for i := range requestQueue {
			request := requestQueue[i]
			go func() {
				c.mutex.Lock()
				fetchers := c.fetchers
				c.mutex.Unlock()
				results := []chan ItemPriceResult{}
				for j := range fetchers {
					result := make(chan ItemPriceResult, 1)
					results = append(results, result)
					go fetchers[j].FetchPrice(request.ctx, request.option, result)
				}

				items := []models.ItemPrice{}
				for j := range results {
					result := <-results[j]
					if result.Err == nil {
						items = append(items, result.Prices...)
					} else {
						fmt.Print("result.Err ", result.Err)
					}
				}

				priceResult := models.MakePriceResult(request.id, items)
				c.mutex.Lock()
				completeHandler := c.completeHandler
				c.mutex.Unlock()
				if completeHandler != nil {
					completeHandler(priceResult)
				}
			}()
		}
	}
}
