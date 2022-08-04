package models

type PriceRequestId = string
type ItemPrice struct {
	Title string `json:"title"`
	Min   string `json:"min"`
	Max   string `json:"max"`
	Url   string `json:"url"`
	Image string `json:"image"`
}

func MakeItemPrice(
	iTitle string,
	iMin string,
	iMax string,
	iUrl string,
	iImage string,
) ItemPrice {
	return ItemPrice{
		Title: iTitle,
		Min:   iMin,
		Max:   iMax,
		Url:   iUrl,
		Image: iImage,
	}
}

type PriceResult struct {
	Id    PriceRequestId `json:"id"`
	Items []ItemPrice    `json:"items"`
}

func MakePriceResult(
	iId PriceRequestId,
	iItems []ItemPrice,
) PriceResult {
	return PriceResult{
		Id:    iId,
		Items: iItems,
	}
}
