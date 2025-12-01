package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"csgoskinflip/server/api"
	"csgoskinflip/server/api/utils/log"
	"csgoskinflip/server/calculator"
	"csgoskinflip/server/collector"
	"csgoskinflip/server/config"
	"csgoskinflip/server/utils/mongo"
	"csgoskinflip/server/utils/request"
)

func main() {
	log.Setup("./log")
	fmt.Print("\n\nStarting...")
	fmt.Println("\n\nBind Address:", config.Config.BindAddress)
	mongo.InitMongoClient(config.Config.MongoConnectionString)
	fmt.Println("Mongo Address:", config.Config.MongoConnectionString)
	fmt.Print("\n\n")
	request.InitClient()

	go analyzeMarket()
	go api.SetupRoutes()

	waitForShutdown()
}

func analyzeMarket() {
	csfloatDataList, skinportDataList, c5gameDataList := collector.Run()
	calculator.Run(csfloatDataList, skinportDataList, c5gameDataList)

	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		csfloatDataList, skinportDataList, c5gameDataList := collector.Run()
		calculator.Run(csfloatDataList, skinportDataList, c5gameDataList)
	}
}

func waitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	fmt.Println("\nStopping...")
}
