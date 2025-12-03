package trades

import (
	"csgoskinflip/server/api/utils/errors"
	"csgoskinflip/server/utils/mongo"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

type TradesResponse struct {
	Trades []TradeItem `json:"trades"`
}

type TradeItem struct {
	Name  string `json:"name"`
	Style string `json:"style"`

	BuyPlattform  string `json:"buy_plattform"`
	BuyLink       string `json:"buy_link"`
	SellPlattform string `json:"sell_plattform"`
	SellLink      string `json:"sell_link"`
	Image         string `json:"image_link"`

	BuyPrice          float64 `json:"buy_price"`
	BuyPriceWithFees  float64 `json:"buy_price_with_fees"`
	SellPrice         float64 `json:"sell_price"`
	SellPriceWithFees float64 `json:"sell_price_with_fees"`

	Profit                  float64 `json:"profit"`
	ProfitAfterWithdrawFrom float64 `json:"profit_after_withdraw_from"`
	ProfitAfterWithdrawTo   float64 `json:"profit_after_withdraw_to"`

	ROI float64 `bson:"roi_percent"`
}

type MongoResponse struct {
	Items []MongoItem `bson:"items"`
}

type MongoItem struct {
	Name  string `bson:"name"`
	Style string `bson:"style"`

	BuyPlattform  string `bson:"buy_plattform"`
	BuyLink       string `bson:"buy_link"`
	SellPlattform string `bson:"sell_plattform"`
	SellLink      string `bson:"sell_link"`
	Image         string `bson:"image_link"`

	BuyPrice          float64 `bson:"buy_price"`
	BuyPriceWithFees  float64 `bson:"buy_price_with_fees"`
	SellPrice         float64 `bson:"sell_price"`
	SellPriceWithFees float64 `bson:"sell_price_with_fees"`

	Profit                  float64 `bson:"profit"`
	ProfitAfterWithdrawFrom float64 `bson:"profit_after_withdraw_from"`
	ProfitAfterWithdrawTo   float64 `bson:"profit_after_withdraw_to"`

	ROI float64 `bson:"roi_percent"`
}

func GetTrades() (*TradesResponse, *errors.Error) {

	mongoResp, err := mongo.FindOne[MongoResponse]("csgoskinflip", "trades", bson.M{})
	if err != nil {
		return nil, errors.NewError(fmt.Sprintf("failed to fetch trades data from mongo: %v", err))
	}

	response := TradesResponse{
		Trades: make([]TradeItem, 0, len(mongoResp.Items)),
	}

	for _, m := range mongoResp.Items {
		tradeItem := TradeItem(m)
		response.Trades = append(response.Trades, tradeItem)
	}

	return &response, nil
}
