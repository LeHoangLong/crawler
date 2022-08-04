package views

import (
	"context"
	"crawler_go/internal/models"
)

type SearchOption struct {
	Keyword string
}

type PriceRequestCompleteHandler = func(prices models.PriceResult)

type PriceRequestControllerI interface {
	SubmitPriceRequest(ctx context.Context, option SearchOption) models.PriceRequestId
	SetCompleteHandler(handler PriceRequestCompleteHandler)
}
