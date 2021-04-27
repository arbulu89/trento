package discover

import (
	"log"
  "io/fs"
	"path"
	"fmt"
  "regexp"
  "path/filepath"

	"github.com/pkg/errors"

  "github.com/spf13/viper"
  "github.com/SUSE/sap_host_exporter/lib/sapcontrol"
)

const SAPInstallationPath string = "/usr/sap"
const SAPIdentifierPattern string = "^[A-Z][A-Z0-9]{2}$"
const SAPInstancePattern string = "^[A-Z]*([0-9]{2})$"

type SAPSystemList []*SAPSystem

type SAPSystem struct {
	webService  sapcontrol.WebService
	SAPInstance *sapcontrol.CurrentSapInstance
	Processes   []*sapcontrol.OSProcess
	Instances   []*sapcontrol.SAPInstance
	Properties  []*sapcontrol.InstanceProperty
}


// Find the installed SAP instances in the /usr/sap folder
func getSapSystems() ([]string, error) {
	var instances = []string{}
	reSAPIdentifier := regexp.MustCompile(SAPIdentifierPattern)

	err := filepath.WalkDir(SAPInstallationPath, func(p string, info fs.DirEntry, err error) error {
		if err != nil {
				log.Print(err.Error())
				return err
		}
		if reSAPIdentifier.MatchString(info.Name()) {
			i, err := getSapInstances(p)
			if err != nil {
				log.Print(err.Error())
			}
			instances = append(instances, i[:]...)
		}
		return nil
	})

	if err != nil {
			return nil, errors.Wrap(err, "Error walking the path")
  }

	return instances, nil
}

// Find the installed SAP instances in the /usr/sap/SID folder
func getSapInstances(sapPath string) ([]string, error) {
	var instances = []string{}
	reSAPInstancer := regexp.MustCompile(SAPInstancePattern)

	err := filepath.WalkDir(sapPath, func(p string, info fs.DirEntry, err error) error {
		if err != nil {
				log.Print(err.Error())
		}
		for _, matches := range reSAPInstancer.FindAllStringSubmatch(info.Name(), -1) {
			instances = append(instances, matches[len(matches)-1])
		}
		return nil
	})

	if err != nil {
			return nil, errors.Wrap(err, "Error walking the path")
  }

	return instances, nil
}

func NewSAPSystemsDiscover() (SAPSystemList, error) {
	var sapSystemList = SAPSystemList{}

	instances, err := getSapSystems()
	if err != nil {
			return nil, errors.Wrap(err, "Error walking the path")
  }

	for _, i := range instances {
		s, err := NewSAPSystemDiscover(i)
		if err != nil {
			return nil, errors.Wrap(err, "Error discovering a SAP instance")
		}
		sapSystemList = append(sapSystemList, &s)
	}
	return sapSystemList, nil
}

func NewSAPSystemDiscover(instNumber string) (SAPSystem, error) {
	var sapSystem = SAPSystem{}

  config := viper.New()
  config.SetDefault("sap-control-uds", path.Join("/tmp", fmt.Sprintf(".sapstream5%s13", instNumber)))
  client := sapcontrol.NewSoapClient(config)
  sapSystem.webService = sapcontrol.NewWebService(client)

  instance, err := sapSystem.webService.GetCurrentInstance()
  if err != nil {
		return sapSystem, errors.Wrap(err, "SAPControl web service error")
  }

	log.Printf("New SAP instance found. %s", instance.String())
	sapSystem.SAPInstance = instance

	processes, err := sapSystem.webService.GetProcessList()
  if err != nil {
		return sapSystem, errors.Wrap(err, "SAPControl web service error")
  }

	sapSystem.Processes = processes.Processes

	instances, err := sapSystem.webService.GetSystemInstanceList()
  if err != nil {
		return sapSystem, errors.Wrap(err, "SAPControl web service error")
  }

	sapSystem.Instances = instances.Instances

	properties, err := sapSystem.webService.GetInstanceProperties()
  if err != nil {
		return sapSystem, errors.Wrap(err, "SAPControl web service error")
  }

	sapSystem.Properties = properties.Properties

  return sapSystem, nil
}
