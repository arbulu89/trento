package web

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/trento-project/trento/internal/consul"

	consultemplateconfig "github.com/hashicorp/consul-template/config"
	"github.com/hashicorp/consul-template/manager"
)

//go:embed frontend/assets
var assetsFS embed.FS

//go:embed templates
var templatesFS embed.FS

//go:embed ansible
var ansibleFS embed.FS

type App struct {
	host string
	port int
	Dependencies
}

type Dependencies struct {
	consul consul.Client
	engine *gin.Engine
}

func DefaultDependencies() Dependencies {
	consulClient, _ := consul.DefaultClient()
	engine := gin.Default()

	return Dependencies{consulClient, engine}
}

// shortcut to use default dependencies
func NewApp(host string, port int) (*App, error) {
	return NewAppWithDeps(host, port, DefaultDependencies())
}

func NewAppWithDeps(host string, port int, deps Dependencies) (*App, error) {
	app := &App{
		Dependencies: deps,
		host:         host,
		port:         port,
	}

	engine := deps.engine
	engine.HTMLRender = NewLayoutRender(templatesFS, "templates/*.tmpl")
	engine.Use(ErrorHandler)
	engine.StaticFS("/static", http.FS(assetsFS))
	engine.GET("/", HomeHandler)
	engine.GET("/hosts", NewHostListHandler(deps.consul))
	engine.GET("/hosts/:name", NewHostHandler(deps.consul))
	engine.GET("/hosts/:name/ha-checks", NewHAChecksHandler(deps.consul))
	engine.GET("/clusters", NewClusterListHandler(deps.consul))
	engine.GET("/clusters/:name", NewClusterHandler(deps.consul))
	engine.GET("/environments", NewEnvironmentListHandler(deps.consul))
	engine.GET("/environments/:env", NewEnvironmentHandler(deps.consul))
	engine.GET("/landscapes", NewLandscapeListHandler(deps.consul))
	engine.GET("/landscapes/:land", NewLandscapeHandler(deps.consul))
	engine.GET("/sapsystems", NewSAPSystemListHandler(deps.consul))
	engine.GET("/sapsystems/:sys", NewSAPSystemHandler(deps.consul))

	apiGroup := engine.Group("/api")
	{
		apiGroup.GET("/ping", ApiPingHandler)
	}

	return app, nil
}

func (a *App) Start() error {
	s := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", a.host, a.port),
		Handler:        a,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	/*
		runner, err := NewTemplateRunner()
		if err != nil {
			return err
		}

		go runner.Start()
	*/

	go startAnsibleTicker()

	return s.ListenAndServe()
}

func (a *App) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	a.engine.ServeHTTP(w, req)
}

const ansibleHostsTemplate = `{{- with node }}
{{- $nodename := .Node.Node }}
[all]
{{- range nodes }}
{{- if ne .Node $nodename }}
{{ .Node }}
{{- end }}
{{- end }}
{{- end }}
{{- range $key, $pairs := tree "trento/v0/clusters/" | byKey }}

[{{ $key }}]
{{- range tree (print "trento/v0/clusters/" $key "/crmmon/Nodes") }}
{{- if .Key | contains "/Name" }}
{{ .Value }}
{{- end }}
{{- end }}
{{- end }}
`

func NewTemplateRunner() (*manager.Runner, error) {
	config := consultemplateconfig.DefaultConfig()
	contents := ansibleHostsTemplate
	destination := path.Join("consul.d", "ansible_hosts")
	*config.Templates = append(
		*config.Templates,
		&consultemplateconfig.TemplateConfig{
			Contents:    &contents,
			Destination: &destination,
		},
	)

	runner, err := manager.NewRunner(config, false)
	if err != nil {
		return nil, errors.Wrap(err, "could not start consul-template")
	}

	return runner, nil
}

func createTempAnsible() error {
	err := os.RemoveAll("consul.d/ansible")
	if err != nil {
		log.Print(err)
		return err
	}

	err = os.Mkdir("consul.d/ansible", 0644)
	if err != nil {
		log.Print(err)
		return err
	}

	err = fs.WalkDir(ansibleFS, "ansible", func(fileName string, dir fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !dir.IsDir() {
			content, err := ansibleFS.ReadFile(fileName)
			if err != nil {
				log.Printf("Error reading file %s", fileName)
				return err
			}
			f, err := os.Create(path.Join("consul.d", fileName))
			if err != nil {
				log.Printf("Error creating file %s", fileName)
				return err
			}
			fmt.Fprintf(f, "%s", content)
		} else {
			os.Mkdir(path.Join("consul.d", fileName), 0644)
		}
		return nil
	})

	return nil
}

func startAnsibleTicker() {
	log.Print("Starting the ansible ticker...")
	createTempAnsible()
	araCallback := exec.Command("python3", "-m", "ara.setup.callback_plugins")
	araCallbackPath, err := araCallback.Output()
	if err != nil {
		log.Println("An error occurred while getting ARA callback plugin path:", err)
	}
	araCallbackPathStr := strings.TrimSpace(string(araCallbackPath))

	araAction := exec.Command("python3", "-m", "ara.setup.action_plugins")
	araActionPath, err := araAction.Output()
	if err != nil {
		log.Println("An error occurred while getting ARA actions plugin path:", err)
	}
	araActionPathStr := strings.TrimSpace(string(araActionPath))

	tick := func() {
		log.Print("Running ansible execution...")
		cmd := exec.Command("ansible-playbook", "consul.d/ansible/main.yaml", "--inventory=consul.d/ansible_hosts")
		cmd.Env = append(os.Environ(), fmt.Sprintf("ANSIBLE_CALLBACK_PLUGINS=%s", araCallbackPathStr))
		cmd.Env = append(os.Environ(), fmt.Sprintf("ANSIBLE_ACTION_PLUGINS=%s", araActionPathStr))
		_, err := cmd.CombinedOutput()
		if err != nil {
			log.Println("An error occurred while running ansible:", err)
		}
		//log.Print(string(result))
	}

	interval := 1 * time.Minute

	repeat(tick, interval)
}

func repeat(tick func(), interval time.Duration) {
	// run the first tick immediately
	tick()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			tick()
		}
	}
}
