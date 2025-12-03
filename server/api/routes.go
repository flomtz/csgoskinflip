package api

import (
	"net/http"

	"csgoskinflip/server/api/handlers"
	"csgoskinflip/server/config"
)

func SetupRoutes() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/status", handlers.GetStatus)
	mux.HandleFunc("GET /api/trades", handlers.GetTrades)
	http.ListenAndServe(config.Config.BindAddress, mux)
}
