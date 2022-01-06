package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func main() {

	boardid, ok := os.LookupEnv("KANBAN_BOARDID")
	if !ok {
		log.Fatal("KANBAN_BOARDID not set")
	}

	apikey, ok := os.LookupEnv("KANBAN_APIKEY")
	if !ok {
		log.Fatal("KANBAN_APIKEY not set")
	}

	subdomain, ok := os.LookupEnv("KANBAN_SUBDOMAIN")
	if !ok {
		log.Fatal("KANBAN_SUBDOMAIN not set")
	}

	format := "json" // json, csv, xml

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

	fmt.Printf("%s", body)

}
