package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

type task struct {
	id           int
	position     int
	size         int
	workflowname string
	columnname   string
	lanename     string
}

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

	m := []map[string]interface{}{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		log.Fatalf("unable to unmarshal response, %v", err)
	}

	tasks := []task{}
	for _, tm := range m {

		// get only cards fron the lane default
		if tm["lanename"] != "Default" {
			continue
		}

		// get only cards fron workflow "Tasks"
		if tm["workflow_name"] != "Tasks" {
			continue
		}

		// atoi taskid
		taskid, err := strconv.Atoi(tm["taskid"].(string))
		if err != nil {
			log.Fatalf("unable to convert taskid to int, %v", err)
		}

		// atoi position
		position, err := strconv.Atoi(tm["position"].(string))
		if err != nil {
			log.Fatalf("unable to convert position to int, %v", err)
		}

		// atoi size
		size := 1
		if tm["size"] != nil {
			size, err = strconv.Atoi(tm["size"].(string))
			if err != nil {
				log.Fatalf("unable to convert size to int, %v", err)
			}
		}

		// columnname
		columnname := tm["columnname"].(string)

		// ignore tasks with "Draft" in the columnname
		if strings.Contains(columnname, "Draft") {
			continue
		}

		// lanename
		lanename := tm["lanename"].(string)

		// workflowname
		workflowname := tm["workflow_name"].(string)

		t := task{
			id:           taskid,
			position:     position,
			size:         size,
			workflowname: workflowname,
			columnname:   columnname,
			lanename:     lanename,
		}

		tasks = append(tasks, t)
		// fmt.Println(tm["taskid"], tm["workflow_name"])
	}

	columnOrder := []string{
		"Backlog",
		"Requested",
		"Doing",
		"Review",
		"Staging",
		"Read for Deploy",
		"Production / Done",
	}

	// sort tasks by columnOrder and position
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].columnname == tasks[j].columnname {
			return tasks[i].position < tasks[j].position
		}
		return sort.SearchStrings(columnOrder, tasks[i].columnname) < sort.SearchStrings(columnOrder, tasks[j].columnname)
	})

	// List result
	for _, t := range tasks {
		fmt.Printf("id: %d, position: %d, size: %d, column: %s, lane: %s\n", t.id, t.position, t.size, t.columnname, t.lanename)
	}

}
