package status

import (
	"csgoskinflip/server/api/utils/errors"
)

type StatusResponse struct {
	Status string `json:"status"`
}

func GetStatus() (*StatusResponse, *errors.Error) {

	response := StatusResponse{
		Status: "success",
	}

	return &response, nil
}
