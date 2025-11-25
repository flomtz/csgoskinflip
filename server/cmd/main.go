package main

import (
	"fmt"

	"csgoskinflip/server/api"
	"csgoskinflip/server/api/utils/log"
	"csgoskinflip/server/config"
	"csgoskinflip/server/utils/mongo"
)

func main() {
	log.Setup("./log")
	fmt.Println("Starting...")
	fmt.Println("Bind Address:", config.Config.BindAddress)
	mongo.InitMongoClient(config.Config.MongoConnectionString)
	fmt.Println("Mongo Address:", config.Config.MongoConnectionString)
	api.SetupRoutes()
}
