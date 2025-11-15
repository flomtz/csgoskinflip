package main

import (
	"fmt"

	"csgoskinflip/server/api"
	"csgoskinflip/server/api/utils/log"
	"csgoskinflip/server/config"
)

func main() {
	log.Setup("./log")
	fmt.Println("Starting...")
	fmt.Println("Bind Address:", config.Config.BindAddress)
	api.SetupRoutes()
}
