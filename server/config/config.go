package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

var Config AppConfig

type AppConfig struct {
	BindAddress           string `json:"bind_address"`
	MongoConnectionString string `json:"mongo_connection_string"`
	C5GameAPIKey          string `json:"c5game_api_key"`
	CsfloatAPIKey         string `json:"csfloat_api_key"`
}

func loadConfig() error {
	workingDir, err := os.Getwd()
	if err != nil {
		return err
	}
	configPath := filepath.Join(workingDir, "config", "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &Config)
}

func init() { // init is a special function that is automatically called
	err := loadConfig()
	if err != nil {
		fmt.Println("Error: Could not load config, using defaults to not break functionality! This is critical!", err)
		Config = AppConfig{BindAddress: ":8080", MongoConnectionString: "mongodb://localhost:27017", C5GameAPIKey: "xxx", CsfloatAPIKey: "xxx"}
	}
}
