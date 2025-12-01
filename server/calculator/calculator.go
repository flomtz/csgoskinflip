package calculator

import (
	"csgoskinflip/server/collector/markets"
	"csgoskinflip/server/utils/mongo"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type TradeItem struct {
	Name  string `bson:"name"`
	Style string `bson:"style"`

	BuyLink  string `bson:"buy_link"`
	SellLink string `bson:"sell_link"`
	Image    string `bson:"image_link"`

	BuyPrice          float64 `bson:"buy_price"`
	BuyPriceWithFees  float64 `bson:"buy_price_with_fees"`
	SellPrice         float64 `bson:"sell_price"`
	SellPriceWithFees float64 `bson:"sell_price_with_fees"`

	Profit                  float64 `bson:"profit"`
	ProfitAfterWithdrawFrom float64 `bson:"profit_after_withdraw_from"`
	ProfitAfterWithdrawTo   float64 `bson:"profit_after_withdraw_to"`
}

type CalculatorOutput struct {
	Trades []TradeItem `json:"trades"`
}

func Run(csfloatDataList []markets.SkinItem, skinportDataList []markets.SkinItem, c5gameDataList []markets.SkinItem) {
	result := &CalculatorOutput{
		Trades: []TradeItem{},
	}

	skinportMap := buildSkinMap(skinportDataList)
	c5gameMap := buildSkinMap(c5gameDataList)

	for _, csfloatSkin := range csfloatDataList {
		csfloatPrice := csfloatSkin.Price * 0.98

		skinportSkin := skinportMap[csfloatSkin.Name]
		c5gameSkin := c5gameMap[csfloatSkin.Name]

		if skinportSkin.Price == 0 && c5gameSkin.Price == 0 {
			continue
		}

		var skinportPrice, skinportProfit float64
		if skinportSkin.Price > 0 {
			skinportPrice = skinportSkin.Price
			skinportProfit = csfloatPrice - skinportPrice
		} else {
			skinportProfit = -999999999999
		}

		var c5gamePrice, c5gameProfit float64
		if c5gameSkin.Price > 0 {
			c5gamePrice = c5gameSkin.Price * 1.01
			c5gameProfit = csfloatPrice - c5gamePrice
		} else {
			c5gameProfit = -999999999999
		}

		if skinportProfit > 0 && skinportProfit >= c5gameProfit {
			result.Trades = append(result.Trades, TradeItem{
				Name:                    csfloatSkin.Name,
				Style:                   csfloatSkin.Style,
				BuyLink:                 skinportSkin.Link,
				SellLink:                "",
				Image:                   "",
				BuyPrice:                skinportSkin.Price,
				BuyPriceWithFees:        skinportPrice,
				SellPrice:               csfloatSkin.Price,
				SellPriceWithFees:       csfloatPrice,
				Profit:                  skinportProfit,
				ProfitAfterWithdrawFrom: (csfloatSkin.Price * 0.955) - skinportPrice,
				ProfitAfterWithdrawTo:   (csfloatSkin.Price * 0.975) - skinportPrice,
			})
		} else if c5gameProfit > 0 {
			result.Trades = append(result.Trades, TradeItem{
				Name:                    csfloatSkin.Name,
				Style:                   csfloatSkin.Style,
				BuyLink:                 c5gameSkin.Link,
				SellLink:                "",
				Image:                   "",
				BuyPrice:                c5gameSkin.Price,
				BuyPriceWithFees:        c5gamePrice,
				SellPrice:               csfloatSkin.Price,
				SellPriceWithFees:       csfloatPrice,
				Profit:                  c5gameProfit,
				ProfitAfterWithdrawFrom: (csfloatSkin.Price * 0.955) - c5gamePrice,
				ProfitAfterWithdrawTo:   (csfloatSkin.Price * 0.975) - c5gamePrice,
			})
		}
	}

	err := saveTradesToMongo(result)
	if err != nil {
		fmt.Printf("failed to save trades data to mongodb: %v\n", err)
	}
}

func buildSkinMap(data []markets.SkinItem) map[string]markets.SkinItem {
	apiMap := make(map[string]markets.SkinItem)
	for _, skin := range data {
		apiMap[skin.Name] = skin
	}
	return apiMap
}

func saveTradesToMongo(data *CalculatorOutput) error {
	database := "csgoskinflip"
	collection := "trades"

	document := bson.M{
		"items": data.Trades,
		"time":  time.Now(),
	}

	exists, err := mongo.CheckForExisting(database, collection, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to check for existing document: %w", err)
	}

	if exists {
		update := bson.M{
			"$set": bson.M{
				"items": data.Trades,
				"time":  time.Now(),
			},
		}
		err = mongo.UpdateOne(database, collection, update, bson.M{})
		if err != nil {
			return fmt.Errorf("failed to update bulk data: %w", err)
		}
	} else {
		err = mongo.InsertOne(database, collection, document)
		if err != nil {
			return fmt.Errorf("failed to insert bulk data: %w", err)
		}
	}

	return nil
}
