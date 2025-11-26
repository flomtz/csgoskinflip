package request

import (
	"net/http"
	"time"
)

var cli *http.Client

func InitClient() *http.Client {
	if cli == nil {
		cli = &http.Client{
			Timeout: 30 * time.Second,
		}
	}
	return cli
}
