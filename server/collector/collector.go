package collector

import (
	"csgoskinflip/server/collector/markets"
	"csgoskinflip/server/utils/request"
	"encoding/json"
	"fmt"
)

func Run() {

	usdExchange, _, err := getExchangePrices()
	if err != nil {
		fmt.Println(err)
	}

	csfloatData, err := markets.CSFloat(usdExchange)
	if err != nil {
		fmt.Println(err)
	}

	_ = createSkinlist(csfloatData)

}

type RateItem struct {
	EUR float64 `json:"EUR"`
}

type ExchangePrice struct {
	Rates RateItem `json:"rates"`
}

func getExchangePrices() (usd float64, rmb float64, err error) {
	respUSD, err := request.Get("https://api.frankfurter.dev/v1/latest?base=USD", nil)
	if err != nil {
		return 0, 0, err
	}

	var USDData ExchangePrice
	err = json.Unmarshal(respUSD.Body, &USDData)
	if err != nil {
		return 0, 0, err
	}

	usdExchange := USDData.Rates.EUR

	respRMB, err := request.Get("https://api.frankfurter.dev/v1/latest?base=CNY", nil)
	if err != nil {
		return 0, 0, err
	}

	var RMBData ExchangePrice
	err = json.Unmarshal(respRMB.Body, &RMBData)
	if err != nil {
		return 0, 0, err
	}

	rmbExchange := RMBData.Rates.EUR

	return usdExchange, rmbExchange, nil
}

type SkinListItem struct {
	Name  string `json:"name"`
	Style string `json:"style"`
}

func createSkinlist(csfloatData *markets.CSFloatOutput) []SkinListItem {
	skinlist := []SkinListItem{}
	csfloatItems := csfloatData.CSFloat

	for _, item := range csfloatItems {
		skinlist = append(skinlist, SkinListItem{
			Name:  item.Name,
			Style: item.Style,
		})
	}

	return skinlist
}
