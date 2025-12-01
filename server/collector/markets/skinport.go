package markets

import (
	"context"
	"csgoskinflip/server/utils/mongo"
	"csgoskinflip/server/utils/request"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type SkinportItem struct {
	Name  string  `json:"name"`
	Style string  `json:"style"`
	Link  string  `json:"link"`
	Price float64 `json:"price"`
}

type SkinportOutput struct {
	Skinport []SkinportItem `json:"skinport"`
}

type SkinportAPIResponse struct {
	MarketHashName string  `json:"market_hash_name"`
	Version        *string `json:"version"`
	ItemPage       string  `json:"item_page"`
	MinPrice       float64 `json:"min_price"`
}

type SkinListItem struct {
	Name  string
	Style string
}

func Skinport(skinList []SkinListItem) error {
	result := &SkinportOutput{
		Skinport: []SkinportItem{},
	}

	apiData, err := fetchPriceList()
	if err != nil {
		return fmt.Errorf("failed to fetch price list: %w", err)
	}

	skinMap := make(map[string]string)
	for _, skin := range skinList {
		skinMap[skin.Name] = skin.Style
	}

	for _, apiItem := range apiData {
		version := ""
		if apiItem.Version != nil {
			version = *apiItem.Version
		}

		name := apiItem.MarketHashName
		style := version

		if expectedStyle, exists := skinMap[name]; exists {
			if expectedStyle == "" || expectedStyle == style {
				result.Skinport = append(result.Skinport, SkinportItem{
					Name:  name,
					Style: style,
					Link:  apiItem.ItemPage,
					Price: apiItem.MinPrice,
				})
			}
		}
	}

	err = saveSkinportToMongo(result)
	if err != nil {
		return fmt.Errorf("failed to save skinport data to mongodb: %w", err)
	}

	return nil
}

func fetchPriceList() ([]SkinportAPIResponse, error) {
	headers := map[string]string{
		"Accept-Encoding": "br, gzip, deflate",
		"Accept":          "application/json",
	}

	resp, err := request.Get("https://api.skinport.com/v1/items?app_id=730&currency=EUR", &request.RequestOptions{
		Headers: headers,
		Timeout: 30 * time.Second,
		Context: context.Background(),
	})
	if err != nil {
		return nil, err
	}

	var items []SkinportAPIResponse
	err = json.Unmarshal(resp.Body, &items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func saveSkinportToMongo(data *SkinportOutput) error {
	database := "csgoskinflip"
	collection := "skinport"

	document := bson.M{
		"items": data.Skinport,
		"time":  time.Now(),
	}

	exists, err := mongo.CheckForExisting(database, collection, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to check for existing document: %w", err)
	}

	if exists {
		update := bson.M{
			"$set": bson.M{
				"items": data.Skinport,
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
