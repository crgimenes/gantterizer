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
	ID           int `json:"id"`
	position     int
	Size         int    `json:"size"`
	WorkflowName string `json:"workflowname"`
	ColumnName   string `json:"columnname"`
	LaneName     string `json:"lanename"`
	Title        string `json:"title"`
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

	workflow := os.Getenv("KANBAN_WORKFLOW")
	lane := os.Getenv("KANBAN_LANE")

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
		if lane != "" && lane != tm["lanename"] {
			continue
		}

		// get only cards fron workflow "Tasks"
		if workflow != "" && workflow != tm["workflow_name"] {
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

		// title
		title := tm["title"].(string)

		t := task{
			ID:           taskid,
			position:     position,
			Size:         size,
			WorkflowName: workflowname,
			ColumnName:   columnname,
			LaneName:     lanename,
			Title:        title,
		}

		tasks = append(tasks, t)

		fmt.Printf("id: %d, position: %d, size: %d, workflowname: %s, columnname: %s, lanename: %s\n",
			t.ID, t.position, t.Size, t.WorkflowName, t.ColumnName, t.LaneName)

	}

	column := []string{
		"Backlog",
		"Requested",
		"Doing",
		"Review",
		"Staging",
		"Read for Deploy",
		"Production / Done",
	}

	// sort tasks as in column array
	sort.Slice(tasks, func(i, j int) bool {
		for k := range column {

			if tasks[i].ColumnName == column[k] && tasks[j].ColumnName == column[k] {
				return tasks[i].position < tasks[j].position
			}

			if tasks[i].ColumnName == column[k] {
				return true
			}

			if tasks[j].ColumnName == column[k] {
				return false
			}

		}

		return false
	})

	// List result
	for i, t := range tasks {
		fmt.Printf("i:%2d │ id: %6d │ column: %-17q │ position: %d\n", i, t.ID, t.ColumnName, t.position)
	}

	type gantt struct {
		taskID int
		size   int
		line   int // increase for each task
		day    int // when the task starts
		title  string
	}

	g := []gantt{}
	taskCount := 0
	maxSimultaneousTask := 3
	line := 0
	startDay := 0
	greatestTaskSize := 0

	for _, t := range tasks {
		if taskCount >= maxSimultaneousTask {
			taskCount = 0
			startDay += greatestTaskSize
			greatestTaskSize = 0
		}

		g = append(g, gantt{
			taskID: t.ID,
			size:   t.Size,
			line:   line,
			day:    startDay,
			title:  t.Title,
		})

		if greatestTaskSize < t.Size {
			greatestTaskSize = t.Size
		}

		taskCount++
		line++
	}

	// List result
	for _, gt := range g {
		taskChart := strings.Repeat("#", gt.size*4)
		taskSpace := strings.Repeat(" ", gt.day*4)
		fmt.Printf("%s %s %d\n", taskSpace, taskChart, gt.taskID)
	}

}
