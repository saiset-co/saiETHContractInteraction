package main

import (
	"io/ioutil"
	"net/http"
	"time"
)

func CallSaiCrypto(url string, request string) string {
	req, err := http.NewRequest("GET", url+request, nil)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	_ = time.AfterFunc(5*time.Second, func() {
		resp.Body.Close()
	})
	body, _ := ioutil.ReadAll(resp.Body)
	return string(body)
}
