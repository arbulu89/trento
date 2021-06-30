package web

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"strings"
)

type HostList struct {
	HostResults []*HostResult `json:"results,omitempty"`
}

type HostResult struct {
	Id       int `json:"id,omitempty"`
	Playbook int `json:"playbook,omitempty"`
	Ok       int `json:"ok,omitempty"`
	Failed   int `json:"failed,omitempty"`
}

type ResultList struct {
	ResultResults []*Result `json:"results,omitempty"`
}

type Result struct {
	Id     int    `json:"id,omitempty"`
	Status string `json:"status,omitempty"`
	TaskId int    `json:"task,omitempty"`
}

type Task struct {
	Id   int      `json:"id,omitempty"`
	Name string   `json:"name,omitempty"`
	Tags []string `tags:"tags:omitempty"`
}

type RecordList struct {
	RecordResults []*RecordItem `json:"results,omitempty"`
}

type RecordItem struct {
	Id       int    `json:"id,omitempty"`
	Key      string `json:"key,omitempty"`
	Playbook int    `json:"playbook,omitempty"`
}

type Record struct {
	Id    int    `json:"id,omitempty"`
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type TrentoCheckResults struct {
	Groups map[string]*TrentoCheckGroup
}

type TrentoCheckGroup struct {
	Id      string
	Name    string
	Results map[string]*TrentoCheckResult
}

type TrentoCheckResult struct {
	Id          string
	Name        string
	Description string
	Status      string
}

func getJson(query string) ([]byte, error) {
	var err error
	resp, err := http.Get(query)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	return body, nil
}

func NewAraHosts(host string) *HostList {
	hostList := &HostList{}

	var err error
	resp, err := getJson(fmt.Sprintf("http://%s:%d/api/v1/hosts?name=%s&order=-created", "10.162.32.181", 8000, host))
	if err != nil {
		log.Print(err)
		return hostList
	}

	err = json.Unmarshal(resp, hostList)
	if err != nil {
		log.Print(err)
		return hostList
	}

	return hostList
}

func normalizeStatus(status string) string {
	switch status {
	case "changed":
		return "PASS"
	case "ok":
		return "PASS"
	default:
		return "FAIL"
	}
}

func (h *HostResult) GetResults() *TrentoCheckResults {
	var checkResults = &TrentoCheckResults{
		Groups: make(map[string]*TrentoCheckGroup),
	}
	var groupId string
	var groupName string

	resultList := NewAraResults(h.Playbook, h.Id)
	for _, r := range resultList.ResultResults {
		//log.Printf("Status: %s", r.Status)
		t := NewAraTask(r.TaskId)
		checkResult := &TrentoCheckResult{
			Status: normalizeStatus(r.Status),
		}
		for _, tItem := range t.Tags {
			if strings.HasPrefix(tItem, "group:") {
				g := strings.Split(tItem, ":")
				groupId = g[1]
				groupName = g[2]
				//log.Printf("Group id: %s . Name: %s", groupId, groupName)
			}

			if tItem == "on_failed:warning" && r.Status == "ignored" {
				checkResult.Status = "WARN"
			}

			if strings.HasPrefix(tItem, "check:") {
				c := strings.Split(tItem, ":")
				//log.Printf("Check id: %s . Name: %s", c[1], c[2])
				checkResult.Id = c[1]
				checkResult.Name = c[2]
			}

			if strings.HasPrefix(tItem, "description:") {
				d := strings.TrimSpace(strings.Split(tItem, "description:")[1])
				//log.Printf("Check description: %s", d)
				checkResult.Description = d
			}
		}

		if checkResult.Id == "" {
			continue
		}

		_, ok := checkResults.Groups[groupId]
		if !ok {
			newGroup := &TrentoCheckGroup{
				Id:      groupId,
				Name:    groupName,
				Results: make(map[string]*TrentoCheckResult),
			}
			newGroup.Results[checkResult.Id] = checkResult
			checkResults.Groups[groupId] = newGroup
		} else {
			checkResults.Groups[groupId].Results[checkResult.Id] = checkResult
		}
	}

	return checkResults
}

func NewAraResults(playbook, host int) *ResultList {
	rList := &ResultList{}

	var err error
	resp, err := getJson(fmt.Sprintf("http://%s:%d/api/v1/results?playbook=%d&host=%d", "10.162.32.181", 8000, playbook, host))
	if err != nil {
		log.Print(err)
		return rList
	}

	err = json.Unmarshal(resp, rList)
	if err != nil {
		log.Print(err)
		return rList
	}

	return rList
}

func NewAraTask(taskId int) *Task {
	rList := &Task{}

	var err error
	resp, err := getJson(fmt.Sprintf("http://%s:%d/api/v1/tasks/%d", "10.162.32.181", 8000, taskId))
	if err != nil {
		log.Print(err)
		return rList
	}

	err = json.Unmarshal(resp, rList)
	if err != nil {
		log.Print(err)
		return rList
	}

	return rList
}

func NewAraRecords(playbook int) *RecordList {
	rList := &RecordList{}

	var err error
	resp, err := getJson(fmt.Sprintf("http://%s:%d/api/v1/records?playbook=%d", "10.162.32.181", 8000, playbook))
	if err != nil {
		log.Print(err)
		return rList
	}

	err = json.Unmarshal(resp, rList)
	if err != nil {
		log.Print(err)
		return rList
	}

	return rList
}

func NewAraRecord(recordId int) *Record {
	r := &Record{}

	var err error
	resp, err := getJson(fmt.Sprintf("http://%s:%d/api/v1/records/%d", "10.162.32.181", 8000, recordId))
	if err != nil {
		log.Print(err)
		return r
	}

	err = json.Unmarshal(resp, r)
	if err != nil {
		log.Print(err)
		return r
	}

	return r
}
