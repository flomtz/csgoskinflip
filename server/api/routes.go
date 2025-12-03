package api

import (
	"net/http"

	"csgoskinflip/server/api/handlers"
	"csgoskinflip/server/config"
)

func SetupRoutes() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/status", handlers.GetStatus)
	http.ListenAndServe(config.Config.BindAddress, mux)
}
