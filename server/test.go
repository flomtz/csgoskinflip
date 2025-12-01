package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func main() {

	url := "https://openapi.c5game.com//merchant/market/v2/products/condition/hash/name?app-key=4a133485df944961906167a297f98bbb"
	method := "POST"

	payload := strings.NewReader(`{
    "pageNum": 1,
    "pageSize": 20,
    "appId": 730,
    "marketHashName": "Sticker | BLAST.tv | Paris 2023",
    "maxPrice": 1000,
    "delivery": 2
}`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}
