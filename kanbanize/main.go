package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {

	const (
		boardid   = "YOUR_BOARD_ID"
		apikey    = "YOUR_API_KEY"
		subdomain = "YOUR_SUBDOMAIN"
		format    = "json" // json, csv, xml
	)

	url := fmt.Sprintf("https://%s.kanbanize.com/index.php/api/kanbanize/get_all_tasks/format/%s", subdomain, format)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(fmt.Sprintf("{\"boardid\":\"%s\"}", boardid))))
	if err != nil {
		log.Fatalf("unable to create request, %v", err)
	}
	req.Header.Add("apikey", apikey)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("request error, %v", err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("unable to read response body, %v", err)
	}

	log.Printf("%s", body)

}
