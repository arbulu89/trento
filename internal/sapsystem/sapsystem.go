package sapsystem

import (
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
	"io/ioutil"

	"github.com/SUSE/sap_host_exporter/lib/sapcontrol"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/spf13/viper"

	"github.com/trento-project/trento/internal"
)

const (
	sapInstallationPath  string = "/usr/sap"
	sapIdentifierPattern string = "^[A-Z][A-Z0-9]{2}$" // PRD, HA1, etc
	sapInstancePattern   string = "^[A-Z]+([0-9]{2})$" // HDB00, ASCS00, ERS10, etc
	sapDefaultProfile    string = "DEFAULT.PFL"
)

const (
	Database = iota + 1
	Application
)

type SAPSystemsList []*SAPSystem

// A SAPSystem in this context is a SAP installation under one SID.
// It will have application or database type, mutually exclusive
// The Id parameter is not yet implemented
type SAPSystem struct {
	//Id         string                `mapstructure:"id,omitempty"`
	SID       string                  `mapstructure:"sid,omitempty"`
	Type      int                     `mapstructure:"type,omitempty"`
	Profile   SAPProfile              `mapstructure:"profile,omitempty"`
	Instances map[string]*SAPInstance `mapstructure:"instances,omitempty"`
}

// The value is interface{} as some of the entries in the SAP profiles files
// are already using "/", so the result will be a map of strings/maps
type SAPProfile map[string]interface{}

type SAPInstance struct {
	Name       string      `mapstructure:"name,omitempty"`
	Type       int         `mapstructure:"type,omitempty"`
	Host       string      `mapstructure:"host,omitempty"`
	SAPControl *SAPControl `mapstructure:"sapcontrol,omitempty"`
}

type SAPControl struct {
	webService sapcontrol.WebService
	Processes  map[string]*sapcontrol.OSProcess        `mapstructure:"processes,omitempty"`
	Instances  map[string]*sapcontrol.SAPInstance      `mapstructure:"instances,omitempty"`
	Properties map[string]*sapcontrol.InstanceProperty `mapstructure:"properties,omitempty"`
}

var newWebService = func(instNumber string) sapcontrol.WebService {
	config := viper.New()
	config.SetDefault("sap-control-uds", path.Join("/tmp", fmt.Sprintf(".sapstream5%s13", instNumber)))
	client := sapcontrol.NewSoapClient(config)
	return sapcontrol.NewWebService(client)
}

var getProfilePath = func(sysPath string) string {
	return path.Join(sysPath, "SYS", "profile", sapDefaultProfile)
}

func NewSAPSystemsList() (SAPSystemsList, error) {
	var systems = SAPSystemsList{}

	appFS := afero.NewOsFs()
	systemPaths, err := findSystems(appFS)
	if err != nil {
		return systems, errors.Wrap(err, "Error walking the path")
	}

	// Find systems
	for _, sysPath := range systemPaths {
		system, err := NewSAPSystem(appFS, sysPath)
		if err != nil {
			log.Printf("Error discovering a SAP system: %s", err)
			continue
		}
		systems = append(systems, system)
	}

	return systems, nil
}

func NewSAPSystem(fs afero.Fs, sysPath string) (*SAPSystem, error) {
	system := &SAPSystem{
		SID:       sysPath[strings.LastIndex(sysPath, "/")+1:],
		Instances: make(map[string]*SAPInstance),
	}

	profilePath := getProfilePath(sysPath)
	profile, err := getProfileData(profilePath)
	if err != nil {
		log.Print(err.Error())
		return system, err
	}
	system.Profile = profile

	instPaths, err := findInstances(fs, sysPath)
	if err != nil {
		log.Print(err.Error())
		return system, err
	}

	// Find instances
	for _, instPath := range instPaths {
		webService := newWebService(instPath[1])
		instance, err := NewSAPInstance(webService)
		if err != nil {
			log.Printf("Error discovering a SAP instance: %s", err)
			continue
		}

		system.Type = instance.Type
		system.Instances[instance.Name] = instance
	}

	return system, nil
}

// Find the installed SAP instances in the /usr/sap folder
// It returns a list of paths where SAP system is found
func findSystems(fs afero.Fs) ([]string, error) {
	var systems = []string{}

	exists, _ := afero.DirExists(fs, sapInstallationPath)
	if !exists {
		log.Print("SAP installation not found")
		return systems, nil
	}

	files, err := afero.ReadDir(fs, sapInstallationPath)
	if err != nil {
		return nil, err
	}

	reSAPIdentifier := regexp.MustCompile(sapIdentifierPattern)

	for _, f := range files {
		if reSAPIdentifier.MatchString(f.Name()) {
			log.Printf("New SAP system installation found: %s", f.Name())
			systems = append(systems, path.Join(sapInstallationPath, f.Name()))
		}
	}

	return systems, nil
}

// Find the installed SAP instances in the /usr/sap/${SID} folder
func findInstances(fs afero.Fs, sapPath string) ([][]string, error) {
	var instances = [][]string{}
	reSAPInstancer := regexp.MustCompile(sapInstancePattern)

	files, err := afero.ReadDir(fs, sapPath)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		for _, matches := range reSAPInstancer.FindAllStringSubmatch(f.Name(), -1) {
			log.Printf("New SAP instance installation found: %s", matches[0])
			instances = append(instances, matches)
		}
	}

	return instances, nil
}

// Get SAP profile file content
func getProfileData(profilePath string) (map[string]interface{}, error) {
	profile, err := os.Open(profilePath)
	if err != nil {
		return nil, fmt.Errorf("could not open profile file %s", err)
	}

	defer profile.Close()

	profileRaw, err := ioutil.ReadAll(profile)

	if err != nil {
		return nil, fmt.Errorf("could not read profile file %s", err)
	}

	configMap := internal.FindMatches(`(\S+)\s=\s(\S+)`, profileRaw)

	return configMap, nil
}

func NewSAPInstance(w sapcontrol.WebService) (*SAPInstance, error) {
	host, _ := os.Hostname()
	var sapInstance = &SAPInstance{
		Host: host,
	}

	scontrol, err := NewSAPControl(w)
	if err != nil {
		return sapInstance, err
	}

	sapInstance.SAPControl = scontrol
	sapInstance.Name = sapInstance.SAPControl.Properties["INSTANCE_NAME"].Value

	_, ok := sapInstance.SAPControl.Properties["HANA Roles"]
	if ok {
		sapInstance.Type = Database
	} else {
		sapInstance.Type = Application
	}

	return sapInstance, nil
}

func NewSAPControl(w sapcontrol.WebService) (*SAPControl, error) {
	var scontrol = &SAPControl{
		webService: w,
		Processes:  make(map[string]*sapcontrol.OSProcess),
		Instances:  make(map[string]*sapcontrol.SAPInstance),
		Properties: make(map[string]*sapcontrol.InstanceProperty),
	}

	properties, err := scontrol.webService.GetInstanceProperties()
	if err != nil {
		return scontrol, errors.Wrap(err, "SAPControl web service error")
	}

	for _, prop := range properties.Properties {
		scontrol.Properties[prop.Property] = prop
	}

	processes, err := scontrol.webService.GetProcessList()
	if err != nil {
		return scontrol, errors.Wrap(err, "SAPControl web service error")
	}

	for _, proc := range processes.Processes {
		scontrol.Processes[proc.Name] = proc
	}

	instances, err := scontrol.webService.GetSystemInstanceList()
	if err != nil {
		return scontrol, errors.Wrap(err, "SAPControl web service error")
	}

	for _, inst := range instances.Instances {
		scontrol.Instances[inst.Hostname] = inst
	}

	return scontrol, nil
}

// This is a unique identifier of a SAP installation.
// It will be used to create totally independent SAP system data
// TODO: This method to obtain the ID must be changed, as this file is not always static
func getUniqueId(sid string) (string, error) {
	return internal.Md5sum(fmt.Sprintf("/usr/sap/%s/SYS/global/security/rsecssfs/key/SSFS_%s.KEY", sid, sid))
}
