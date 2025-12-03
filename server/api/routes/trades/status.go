package trades

import (
	"csgoskinflip/server/api/utils/errors"
)

type TradesResponse struct {
}

func GetTrades() (*TradesResponse, *errors.Error) {

	response := TradesResponse{}

	return &response, nil
}
