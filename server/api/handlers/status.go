package handlers

import (
	"net/http"

	"csgoskinflip/server/api/routes/status"
	"csgoskinflip/server/api/utils/errors"
	"csgoskinflip/server/api/utils/responses"
)

func GetStatus(w http.ResponseWriter, r *http.Request) {
	responses.HandleRequest(w, r, func() (*status.StatusResponse, *errors.Error) {
		return status.GetStatus()
	})
}
