package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

var Config AppConfig

type AppConfig struct {
	BindAddress string `json:"bind_address"`
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
		Config = AppConfig{BindAddress: ":8080"}
	}
}
