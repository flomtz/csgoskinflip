package cache

import (
	"context"
	"csgoskinflip/server/collector/markets"
	"csgoskinflip/server/config"
	"csgoskinflip/server/utils/mongo"
	"csgoskinflip/server/utils/request"
	"encoding/json"
	"fmt"
	"maps"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type CacheItem struct {
	DefIndex   int    `bson:"def_index"`
	PaintIndex int    `bson:"paint_index"`
	ImageURL   string `bson:"image_url"`
	WearName   string `bson:"wear_name"`
}

type CacheData struct {
	Map map[string]CacheItem `bson:"cache"`
}

func SkinCache(skinlist []markets.SkinListItem) (map[string]CacheItem, error) {
	exists, err := mongo.CollectionExists("csgoskinflip", "cache")
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing collection: %v", err)
	}

	if !exists {
		document := bson.M{
			"cache": map[string]CacheItem{},
		}

		err = mongo.InsertOne("csgoskinflip", "cache", document)
		if err != nil {
			return nil, fmt.Errorf("failed initialize cache collection in mongo db: %v", err)
		}
	}

	cacheData, err := mongo.FindOne[CacheData]("csgoskinflip", "cache", bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch cache data from mongo: %v", err)
	}

	skinCacheMap := cacheData.Map
	var toFetch []markets.SkinListItem
	var name string

	for _, skin := range skinlist {

		if skin.Style == "" {
			name = skin.Name
		} else {
			name = fmt.Sprintf("%s %s", skin.Name, skin.Style)
		}

		_, exists := skinCacheMap[name]
		if exists {
			continue
		} else {
			toFetch = append(toFetch, skin)
		}
	}

	if len(toFetch) != 0 {
		newCacheDataMap, err := fetchCsfloatData(toFetch)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch new csfloat data: %v", err)
		}

		maps.Copy(skinCacheMap, newCacheDataMap)

		result := &CacheData{
			Map: skinCacheMap,
		}

		err = saveCacheToMongo(result)
		if err != nil {
			return nil, fmt.Errorf("failed to save csfloat data to mongodb: %w", err)
		}
	}

	return skinCacheMap, nil
}

func fetchCsfloatData(skinlist []markets.SkinListItem) (map[string]CacheItem, error) {
	cacheMap := make(map[string]CacheItem)
	var name string

	for _, skin := range skinlist {
		defIndex, paintIndex, imageURL, wearName, err := getSingleSkinData(skin.Name, skin.Style)
		if err != nil {
			return nil, fmt.Errorf("failed to get skin data: %v", err)
		}

		if skin.Style == "" {
			name = skin.Name
		} else {
			name = fmt.Sprintf("%s %s", skin.Name, skin.Style)
		}

		cacheMap[name] = CacheItem{
			DefIndex:   defIndex,
			PaintIndex: paintIndex,
			ImageURL:   imageURL,
			WearName:   wearName,
		}
	}

	return cacheMap, nil
}

type CsfloatResponse struct {
	Data []Listing `json:"data"`
}

type Listing struct {
	Item ItemInfo `json:"item"`
}

type ItemInfo struct {
	DefIndex   int    `json:"def_index"`
	PaintIndex int    `json:"paint_index"`
	ImageURL   string `json:"icon_url"`
	Phase      string `json:"phase"`
	WearName   string `json:"wear_name"`
}

func getSingleSkinData(name string, style string) (int, int, string, string, error) {
	apiKey := config.Config.CsfloatAPIKey
	apiURL := "https://csfloat.com/api/v1/listings"

	var defIndex int
	var paintIndex int
	var imageURL string
	var wearName string

	if style == "" {
		resp, err := request.Get(apiURL, &request.RequestOptions{
			Context: context.Background(),
			Timeout: 30 * time.Second,
			Headers: map[string]string{
				"Authorization": apiKey,
			},
			QueryParams: map[string]string{
				"market_hash_name": name,
				"limit":            "1",
				"type":             "buy_now",
			},
		})
		if err != nil {
			return 0, 0, "", "", fmt.Errorf("failed to call csfloat api: %v", err)
		}

		var apiResp CsfloatResponse
		err = json.Unmarshal(resp.Body, &apiResp)
		if err != nil {
			return 0, 0, "", "", fmt.Errorf("failed to call csfloat api: %v", err)
		}

		if len(apiResp.Data) == 0 {
			defIndex = 0
			paintIndex = 0
			imageURL = ""
			wearName = ""
		} else {
			bestListing := apiResp.Data[0]
			defIndex = bestListing.Item.DefIndex
			paintIndex = bestListing.Item.PaintIndex
			imageURL = bestListing.Item.ImageURL
			wearName = bestListing.Item.WearName
		}

	} else {
		resp, err := request.Get(apiURL, &request.RequestOptions{
			Context: context.Background(),
			Timeout: 30 * time.Second,
			Headers: map[string]string{
				"Authorization": apiKey,
			},
			QueryParams: map[string]string{
				"market_hash_name": name,
				"limit":            "50",
				"type":             "buy_now",
			},
		})
		if err != nil {
			return 0, 0, "", "", fmt.Errorf("failed to call csfloat api: %v", err)
		}

		var apiResp CsfloatResponse
		err = json.Unmarshal(resp.Body, &apiResp)
		if err != nil {
			return 0, 0, "", "", fmt.Errorf("failed to call csfloat api: %v", err)
		}

		for _, data := range apiResp.Data {
			if data.Item.Phase == style {
				defIndex = data.Item.DefIndex
				paintIndex = data.Item.PaintIndex
				imageURL = data.Item.ImageURL
				break
			} else {
				continue
			}
		}
	}

	return defIndex, paintIndex, imageURL, wearName, nil
}

func saveCacheToMongo(data *CacheData) error {
	database := "csgoskinflip"
	collection := "cache"

	document := bson.M{
		"cache": data.Map,
		"time":  time.Now(),
	}

	exists, err := mongo.CheckForExisting(database, collection, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to check for existing document: %w", err)
	}

	if exists {
		update := bson.M{
			"$set": bson.M{
				"cache": data.Map,
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
