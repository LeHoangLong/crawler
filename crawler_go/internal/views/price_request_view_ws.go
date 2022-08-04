package views

import (
	"context"
	"crawler_go/internal/models"
	"crawler_go/internal/server"
	"fmt"
	"sync"
)

type PriceRequestViewWs struct {
	controller     PriceRequestControllerI
	reponseWriters map[models.PriceRequestId]server.WsResponseWriter
	mutex          *sync.Mutex
}

type WsPriceRequest struct {
	Keyword string `json:"keyword"`
}

func MakePriceRequestViewWs(
	iController PriceRequestControllerI,
) PriceRequestViewWs {
	view := PriceRequestViewWs{
		controller:     iController,
		mutex:          &sync.Mutex{},
		reponseWriters: map[models.PriceRequestId]server.WsResponseWriter{},
	}
	iController.SetCompleteHandler(view.PriceRequestCompleteHandler)

	return view
}

func (v *PriceRequestViewWs) PriceRequestCompleteHandler(priceResult models.PriceResult) {
	v.mutex.Lock()
	if response, ok := v.reponseWriters[priceResult.Id]; ok {
		delete(v.reponseWriters, priceResult.Id)
		v.mutex.Unlock()
		response.SendData(priceResult)
	} else {
		v.mutex.Unlock()
	}
}

func (v *PriceRequestViewWs) HandleWsRequest(ctx context.Context, request map[string]interface{}, response server.WsResponseWriter) error {
	fmt.Println("start handle ws request")
	if keyword, ok := request["keyword"]; !ok {
		response.SendError(server.InvalidFormat)
	} else {
		if keywordStr, ok := keyword.(string); !ok {
			response.SendError(server.InvalidFormat)
		} else {
			option := SearchOption{
				Keyword: keywordStr,
			}
			id := v.controller.SubmitPriceRequest(ctx, option)
			v.mutex.Lock()
			v.reponseWriters[id] = response
			v.mutex.Unlock()
		}
	}
	return nil
}
