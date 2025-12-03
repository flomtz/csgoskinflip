package handlers

import (
	"net/http"

	"csgoskinflip/server/api/routes/trades"
	"csgoskinflip/server/api/utils/errors"
	"csgoskinflip/server/api/utils/responses"
)

func GetTrades(w http.ResponseWriter, r *http.Request) {
	responses.HandleRequest(w, r, func() (*trades.TradesResponse, *errors.Error) {
		return trades.GetTrades()
	})
}
