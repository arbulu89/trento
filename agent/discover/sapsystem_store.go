package discover

import (
	"log"
	"fmt"
	"strings"
	"strconv"

	"github.com/pkg/errors"

	consulApi "github.com/hashicorp/consul/api"
	"github.com/trento-project/trento/internal/consul"
	"github.com/SUSE/sap_host_exporter/lib/sapcontrol"
)

const (
  sapSystemPath string = "trento/nodes/%s/sapsystems/%s"
  name          string = "name"
  sidPath       string = "sid"
  instancePath  string = "instance"
  processesPath string = "processes"
)

func (s *SAPSystem) getKVPath() string {
  host := s.SAPInstance.Hostname
  name := s.SAPInstance.Name
  kvPath := fmt.Sprintf(sapSystemPath, host, name)

  return kvPath
}

func (s *SAPSystem) storeProcesses(client consul.Client) error {
  for _, p := range s.Processes {
    kvPath := fmt.Sprintf("%s/%s/%s", s.getKVPath(), processesPath, p.Name)

    err := store(client, fmt.Sprintf("%s/%s", kvPath, "description"), p.Description)
  	if err != nil {
  		return errors.Wrap(err, "Error storing the SAP process description")
  	}

    err = store(client, fmt.Sprintf("%s/%s", kvPath, "textstatus"), p.Textstatus)
  	if err != nil {
  		return errors.Wrap(err, "Error storing the SAP textstatus name")
  	}

    err = store(client, fmt.Sprintf("%s/%s", kvPath, "starttime"), p.Starttime)
  	if err != nil {
  		return errors.Wrap(err, "Error storing the SAP starttime name")
  	}

    err = store(client, fmt.Sprintf("%s/%s", kvPath, "elapsedtime"), p.Elapsedtime)
  	if err != nil {
  		return errors.Wrap(err, "Error storing the SAP elapsedtime name")
  	}

    err = store(client, fmt.Sprintf("%s/%s", kvPath, "pid"), fmt.Sprintf("%d", p.Pid))
  	if err != nil {
  		return errors.Wrap(err, "Error storing the SAP pid name")
  	}
  }

	return nil
}

func (s *SAPSystem) StoreDiscovery(client consul.Client) error {
  kvPath := s.getKVPath()

	_, err := client.KV().DeleteTree(kvPath, nil)
	if err != nil {
		return errors.Wrap(err, "Error deleting SAP system content")
	}

  err = store(client, fmt.Sprintf("%s/%s", kvPath, sidPath), s.SAPInstance.SID)
	if err != nil {
		return errors.Wrap(err, "Error storing the SAP sid value")
	}

  err = store(client, fmt.Sprintf("%s/%s", kvPath, instancePath), fmt.Sprintf("%02d", s.SAPInstance.Number))
	if err != nil {
		return errors.Wrap(err, "Error storing the SAP sid value")
	}

  s.storeProcesses(client)
  if err != nil {
		return errors.Wrap(err, "Error storing the SAP process")
	}

	return nil
}

func store(client consul.Client, key string, value string) error {
  _, err := client.KV().Put(&consulApi.KVPair{
    Key:   key,
    Value: []byte(value)}, nil)

  if err != nil {
		return errors.Wrap(err, "Error storing a new value in the KV storage")
	}

  log.Printf("Value %s properly stored at %s", value, key)
  return nil
}

func LoadDiscovery(client consul.Client, host string) (map[string]*SAPSystem, error) {
	var sapSystems = map[string]*SAPSystem{}
	kvPath := fmt.Sprintf("trento/nodes/%s/sapsystems", host)
	var lastProcess string

	entries, _, err := client.KV().List(kvPath, nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not query Consul for SAP systems KV values")
	}

	for _, entry := range entries {
		keyValues := strings.Split(strings.TrimSuffix(entry.Key, "/"), "/")
		name := keyValues[4]

		_, found := sapSystems[name]
		if !found {
			sapSystems[name] = &SAPSystem{
				SAPInstance: &sapcontrol.CurrentSapInstance{Name: name},
			}
		}

		if strings.HasSuffix(entry.Key, "sid") {
			sapSystems[name].SAPInstance.SID = string(entry.Value)
		}

		if strings.HasSuffix(entry.Key, "instance") {
			i, _ := strconv.ParseInt(string(entry.Value), 10, 32)
			sapSystems[name].SAPInstance.Number = int32(i)
		}

		if contains(keyValues, "processes") {
			currentProcess := keyValues[6]
			if currentProcess != lastProcess {
				process := &sapcontrol.OSProcess{Name: currentProcess}
				sapSystems[name].Processes = append(sapSystems[name].Processes, process)
			}

			switch keyValues[len(keyValues)-1] {
			case "description":
				sapSystems[name].Processes[len(sapSystems[name].Processes)-1].Description = string(entry.Value)
			case "elapsedtime":
				sapSystems[name].Processes[len(sapSystems[name].Processes)-1].Elapsedtime = string(entry.Value)
			case "pid":
				p, _ := strconv.ParseInt(string(entry.Value), 10, 32)
				sapSystems[name].Processes[len(sapSystems[name].Processes)-1].Pid = int32(p)
			case "starttime":
				sapSystems[name].Processes[len(sapSystems[name].Processes)-1].Starttime = string(entry.Value)
			case "textstatus":
				sapSystems[name].Processes[len(sapSystems[name].Processes)-1].Textstatus = string(entry.Value)

			}
			lastProcess = currentProcess
		}
	}

	return sapSystems, nil
}


func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
