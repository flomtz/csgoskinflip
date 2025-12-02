package calculator

import (
	"csgoskinflip/server/collector/markets"
	"csgoskinflip/server/utils/mongo"
	"fmt"
	"math"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type TradeItem struct {
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
		if skinportSkin.Price > 1 {
			skinportPrice = skinportSkin.Price
			skinportProfit = math.Round((csfloatPrice-skinportPrice)*100) / 100
		} else {
			skinportProfit = -999999999999
		}

		var c5gamePrice, c5gameProfit float64
		if c5gameSkin.Price > 1 {
			c5gamePrice = c5gameSkin.Price * 1.01
			c5gameProfit = math.Round((csfloatPrice-c5gamePrice)*100) / 100
		} else {
			c5gameProfit = -999999999999
		}

		if skinportProfit > 0 && skinportProfit >= c5gameProfit {
			ProfitAfterWithdrawFrom := math.Round(((csfloatSkin.Price*0.955)-skinportPrice)*100) / 100
			ProfitAfterWithdrawTo := math.Round(((csfloatSkin.Price*0.975)-skinportPrice)*100) / 100

			if (ProfitAfterWithdrawFrom <= 0 || ProfitAfterWithdrawTo <= 0) || (csfloatPrice/skinportPrice) > 15 {
				continue
			}

			ROI := math.Round(((csfloatPrice/skinportPrice)*100)*100) / 100

			if ROI < 110 {
				continue
			}

			result.Trades = append(result.Trades, TradeItem{
				Name:                    csfloatSkin.Name,
				Style:                   csfloatSkin.Style,
				BuyPlattform:            "skinport",
				BuyLink:                 skinportSkin.Link,
				SellPlattform:           "csfloat",
				SellLink:                "",
				Image:                   "",
				BuyPrice:                skinportSkin.Price,
				BuyPriceWithFees:        skinportPrice,
				SellPrice:               csfloatSkin.Price,
				SellPriceWithFees:       csfloatPrice,
				Profit:                  skinportProfit,
				ProfitAfterWithdrawFrom: ProfitAfterWithdrawFrom,
				ProfitAfterWithdrawTo:   ProfitAfterWithdrawTo,
				ROI:                     ROI,
			})
		} else if c5gameProfit > 0 {
			ProfitAfterWithdrawFrom := math.Round(((csfloatSkin.Price*0.955)-c5gamePrice)*100) / 100
			ProfitAfterWithdrawTo := math.Round(((csfloatSkin.Price*0.975)-c5gamePrice)*100) / 100

			if (ProfitAfterWithdrawFrom <= 0 || ProfitAfterWithdrawTo <= 0) || (csfloatPrice/c5gamePrice) > 15 {
				continue
			}

			ROI := math.Round(((csfloatPrice/c5gamePrice)*100)*100) / 100

			if ROI < 110 {
				continue
			}

			result.Trades = append(result.Trades, TradeItem{
				Name:                    csfloatSkin.Name,
				Style:                   csfloatSkin.Style,
				BuyPlattform:            "c5game",
				BuyLink:                 c5gameSkin.Link,
				SellPlattform:           "csfloat",
				SellLink:                "",
				Image:                   "",
				BuyPrice:                c5gameSkin.Price,
				BuyPriceWithFees:        c5gamePrice,
				SellPrice:               csfloatSkin.Price,
				SellPriceWithFees:       csfloatPrice,
				Profit:                  c5gameProfit,
				ProfitAfterWithdrawFrom: ProfitAfterWithdrawFrom,
				ProfitAfterWithdrawTo:   ProfitAfterWithdrawTo,
				ROI:                     ROI,
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
