package markets

import (
	"csgoskinflip/server/utils/mongo"
	"csgoskinflip/server/utils/request"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

const (
	csfloatNormalAPI  = "https://csfloat.com/api/v1/listings/price-list"
	csfloatDopplerAPI = "https://csfloat.com/api/v1/listings/price-list/doppler"
)

type NormalPriceItem struct {
	MarketHashName string  `json:"market_hash_name"`
	MinPrice       float64 `json:"min_price"`
}

type DopplerPriceItem struct {
	MarketHashName string  `json:"market_hash_name"`
	MinPrice       float64 `json:"min_price"`
	PhaseName      string  `json:"phase_name"`
}

type DopplerResponse struct {
	Data []DopplerPriceItem `json:"data"`
}

type CSFloatItem struct {
	Name  string  `json:"name"`
	Style string  `json:"style"`
	Price float64 `json:"price"`
}

type CSFloatOutput struct {
	CSFloat []CSFloatItem `json:"csfloat"`
}

func CSFloat(usdExchange float64) (*CSFloatOutput, error) {
	result := &CSFloatOutput{
		CSFloat: []CSFloatItem{},
	}

	normalItems, err := fetchNormalPriceList()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch normal price list: %w", err)
	}

	dopplerItems, err := fetchDopplerPriceList()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch doppler price list: %w", err)
	}

	for _, item := range normalItems {
		result.CSFloat = append(result.CSFloat, CSFloatItem{
			Name:  item.MarketHashName,
			Style: "",
			Price: math.Round(((item.MinPrice/100)*usdExchange)*100) / 100,
		})
	}

	for _, item := range dopplerItems {
		result.CSFloat = append(result.CSFloat, CSFloatItem{
			Name:  item.MarketHashName,
			Style: item.PhaseName,
			Price: math.Round(((item.MinPrice/100)*usdExchange)*100) / 100,
		})
	}

	err = saveToMongo(result)
	if err != nil {
		return nil, fmt.Errorf("failed to save csfloat data to mongodb: %w", err)
	}

	return result, nil
}

func fetchNormalPriceList() ([]NormalPriceItem, error) {
	resp, err := request.Get(csfloatNormalAPI, nil)
	if err != nil {
		return nil, err
	}

	var items []NormalPriceItem
	err = json.Unmarshal(resp.Body, &items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func fetchDopplerPriceList() ([]DopplerPriceItem, error) {
	resp, err := request.Get(csfloatDopplerAPI, nil)
	if err != nil {
		return nil, err
	}

	var response DopplerResponse
	err = json.Unmarshal(resp.Body, &response)
	if err != nil {
		return nil, err
	}

	return response.Data, nil
}

func saveToMongo(data *CSFloatOutput) error {
	database := "csgoskinflip"
	collection := "csfloat"

	document := bson.M{
		"items": data.CSFloat,
		"time":  time.Now(),
	}

	exists, err := mongo.CheckForExisting(database, collection, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to check for existing document: %w", err)
	}

	if exists {
		update := bson.M{
			"$set": bson.M{
				"items": data.CSFloat,
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
