package markets

import (
	"context"
	"csgoskinflip/server/config"
	"csgoskinflip/server/utils/mongo"
	"csgoskinflip/server/utils/request"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type C5GameItem struct {
	Name  string  `json:"name"`
	Style string  `json:"style"`
	Link  string  `json:"link"`
	Price float64 `json:"price"`
}

type C5GameOutout struct {
	C5Game []C5GameItem `json:"c5game"`
}

func C5Game(skinlist []SkinListItem, rmbExchange float64) error {
	_ = &C5GameOutout{
		C5Game: []C5GameItem{},
	}

	start := time.Now()
	var normalSkins []SkinListItem
	var dopplerSkins []SkinListItem

	for _, skin := range skinlist {
		if skin.Style == "" {
			normalSkins = append(normalSkins, SkinListItem{
				Name:  skin.Name,
				Style: skin.Style,
			})
		} else {
			dopplerSkins = append(dopplerSkins, SkinListItem{
				Name:  skin.Name,
				Style: skin.Style,
			})
		}
	}

	_, err := fetchDopplerSkins(dopplerSkins, rmbExchange)
	if err != nil {
		return fmt.Errorf("failed to get doppler skins prices: %v", err)
	}

	fmt.Println("Total time", time.Since(start))

	return nil
}

func fetchNormalSkins(skinlist []SkinListItem, rmbExchange float64) ([]C5GameItem, error) {
	start := time.Now()
	var result []C5GameItem
	apiBulkUrl := fmt.Sprintf("https://openapi.c5game.com/merchant/product/price/batch?app-key=%s", config.Config.C5GameAPIKey)

	bulkList := make([]string, 0, len(skinlist))
	for _, skin := range skinlist {
		bulkList = append(bulkList, skin.Name)
	}

	const batchSize = 200
	const workers = 5

	type batchJob struct {
		start int
		end   int
		items []string
	}

	jobs := make(chan batchJob, workers)
	errors := make(chan error, len(bulkList)/batchSize+10)

	var wg sync.WaitGroup

	type apiResponse struct {
		Success bool `json:"success"`
		Data    map[string]struct {
			MarketHashName string  `json:"marketHashName"`
			Price          float64 `json:"price"`
			Website        string  `json:"website"`
		} `json:"data"`
	}

	var mu sync.Mutex

	worker := func() {
		defer wg.Done()

		for job := range jobs {
			payload := map[string]interface{}{
				"appId":           730,
				"marketHashNames": job.items,
			}

			resp, err := request.Post(apiBulkUrl, payload, &request.RequestOptions{
				Context: context.Background(),
				Timeout: 15 * time.Second,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			})
			if err != nil {
				errors <- fmt.Errorf("batch %d-%d failed: %w", job.start, job.end, err)
				continue
			}

			var apiResp apiResponse
			err = json.Unmarshal(resp.Body, &apiResp)
			if err != nil {
				errors <- fmt.Errorf("invalid JSON in batch %d-%d: %w", job.start, job.end, err)
				continue
			}

			if !apiResp.Success {
				errors <- fmt.Errorf("api error in batch %d-%d", job.start, job.end)
				continue
			}

			mu.Lock()
			for _, item := range apiResp.Data {
				result = append(result, C5GameItem{
					Name:  item.MarketHashName,
					Link:  item.Website,
					Price: math.Round((item.Price*rmbExchange)*100) / 100,
				})
			}
			mu.Unlock()
		}
	}

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go worker()
	}

	for i := 0; i < len(bulkList); i += batchSize {
		end := i + batchSize
		if end > len(bulkList) {
			end = len(bulkList)
		}

		jobs <- batchJob{
			start: i,
			end:   end,
			items: bulkList[i:end],
		}
	}

	close(jobs)
	wg.Wait()
	close(errors)

	errorCount := 0
	for err := range errors {
		fmt.Println("Error:", err)
		errorCount++
	}

	fmt.Println("Errors:", errorCount)
	fmt.Println("Normalskins completed in:", time.Since(start))

	return result, nil
}

var styleMap = map[string]int{
	"":           0,
	"Phase 1":    11,
	"Phase 2":    12,
	"Phase 3":    13,
	"Phase 4":    14,
	"Ruby":       31,
	"Sapphire":   33,
	"Blackpearl": 34,
	"Emerald":    32,
	"Singleblue": 35,
}

func fetchDopplerSkins(skinlist []SkinListItem, rmbExchange float64) ([]C5GameItem, error) {
	start := time.Now()
	var result []C5GameItem
	apiUrl := fmt.Sprintf("https://openapi.c5game.com/merchant/market/v2/products/condition/hash/name?app-key=%s", config.Config.C5GameAPIKey)

	const workers = 5

	errors := make(chan error, len(skinlist))
	var wg sync.WaitGroup
	var mu sync.Mutex

	type apiResponse struct {
		Success bool `json:"success"`
		Data    struct {
			List []struct {
				Price          float64 `json:"price"`
				MarketHashName string  `json:"marketHashName"`
				ItemID         int     `json:"itemId"`
			} `json:"list"`
		} `json:"data"`
	}

	fmt.Println("Skinlist length:", len(skinlist))

	worker := func(skins <-chan SkinListItem) {
		defer wg.Done()

		for skin := range skins {
			styleId := styleMap[skin.Style]
			payload := map[string]interface{}{
				"appId":           730,
				"marketHashNames": skin.Name,
				"styleId":         styleId,
			}

			resp, err := request.Post(apiUrl, payload, &request.RequestOptions{
				Context: context.Background(),
				Timeout: 15 * time.Second,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			})
			if err != nil {
				errors <- fmt.Errorf("%s %s failed: %w", skin.Name, skin.Style, err)
				continue
			}

			var apiResp apiResponse
			err = json.Unmarshal(resp.Body, &apiResp)
			if err != nil {
				errors <- fmt.Errorf("invalid JSON in %s %s: %w", skin.Name, skin.Style, err)
				continue
			}

			if !apiResp.Success {
				errors <- fmt.Errorf("api error in %s %s", skin.Name, skin.Style)
				continue
			}

			if len(apiResp.Data.List) == 0 {
				errors <- fmt.Errorf("no listings for %s %s", skin.Name, skin.Style)
				continue
			}

			bestListing := apiResp.Data.List[0]

			link := fmt.Sprintf(
				"https://www.c5game.com/en/csgo/%d/%s/sell?levelIds=%d",
				bestListing.ItemID,
				url.QueryEscape(bestListing.MarketHashName),
				styleId,
			)

			mu.Lock()
			result = append(result, C5GameItem{
				Name:  bestListing.MarketHashName,
				Link:  link,
				Price: math.Round((bestListing.Price*rmbExchange)*100) / 100,
			})
			mu.Unlock()
		}
	}

	skinChan := make(chan SkinListItem, len(skinlist))
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go worker(skinChan)
	}

	for _, skin := range skinlist {
		skinChan <- skin
	}
	close(skinChan)

	wg.Wait()
	close(errors)

	errorCount := 0
	for err := range errors {
		fmt.Println("Error:", err)
		errorCount++
	}

	fmt.Println("Errors:", errorCount)
	fmt.Println("Dopplerskins completed in:", time.Since(start))

	return result, nil
}

func saveC5GameToMongo(data *C5GameOutout) error {
	database := "csgoskinflip"
	collection := "c5game"

	document := bson.M{
		"items": data.C5Game,
		"time":  time.Now(),
	}

	exists, err := mongo.CheckForExisting(database, collection, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to check for existing document: %w", err)
	}

	if exists {
		update := bson.M{
			"$set": bson.M{
				"items": data.C5Game,
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
