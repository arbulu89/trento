package web

import (
	"net/http"
	"sort"
	"log"

	"github.com/gin-gonic/gin"
	consulApi "github.com/hashicorp/consul/api"
	"github.com/pkg/errors"

	"github.com/trento-project/trento/internal/consul"
	"github.com/trento-project/trento/internal/environments"
	"github.com/trento-project/trento/internal/hosts"
	"github.com/trento-project/trento/internal/sapsystem"
)

func NewHostListHandler(client consul.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Request.URL.Query()
		query_filter := hosts.CreateFilterMetaQuery(query)
		health_filter := query["health"]

		hosts, err := hosts.Load(client, query_filter, health_filter)
		if err != nil {
			_ = c.Error(err)
			return
		}

		filters, err := loadFilters(client)
		if err != nil {
			_ = c.Error(err)
			return
		}

		c.HTML(http.StatusOK, "hosts.html.tmpl", gin.H{
			"Hosts":          hosts,
			"Filters":        filters,
			"AppliedFilters": query,
		})
	}
}

func loadFilters(client consul.Client) (map[string][]string, error) {
	filter_data := make(map[string][]string)

	envs, err := environments.Load(client)
	if err != nil {
		return nil, errors.Wrap(err, "could not get the filters")
	}

	for envKey, envValue := range envs {
		filter_data["environments"] = append(filter_data["environments"], envKey)
		for landKey, landValue := range envValue.Landscapes {
			filter_data["landscapes"] = append(filter_data["landscapes"], landKey)
			for sysKey, _ := range landValue.SAPSystems {
				filter_data["sapsystems"] = append(filter_data["sapsystems"], sysKey)
			}
		}
	}

	sort.Strings(filter_data["environments"])
	sort.Strings(filter_data["landscapes"])
	sort.Strings(filter_data["sapsystems"])

	return filter_data, nil
}

func loadHealthChecks(client consul.Client, node string) ([]*consulApi.HealthCheck, error) {

	checks, _, err := client.Health().Node(node, nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not query Consul for health checks")
	}

	return checks, nil
}

func NewHostHandler(client consul.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		catalogNode, _, err := client.Catalog().Node(name, nil)
		if err != nil {
			_ = c.Error(err)
			return
		}

		if catalogNode == nil {
			_ = c.Error(NotFoundError("could not find host"))
			return
		}

		checks, err := loadHealthChecks(client, name)
		if err != nil {
			_ = c.Error(err)
			return
		}

		systems, err := sapsystem.Load(client, name)
		if err != nil {
			_ = c.Error(err)
			return
		}
		host := hosts.NewHost(*catalogNode.Node, client)
		c.HTML(http.StatusOK, "host.html.tmpl", gin.H{
			"Host":         &host,
			"HealthChecks": checks,
			"SAPSystems":   systems,
		})
	}
}

func NewHAChecksHandler(client consul.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")

		n := NewAraHosts(name)

		// Avoid ongoing test results
		lastResult := n.HostResults[0]
		if lastResult.Ok == 0 || lastResult.Failed == 0 {
			lastResult = n.HostResults[1]
		}

		r := lastResult.GetResults()
		for _, rItem := range r.Groups {
			log.Print(rItem)
		}

		catalogNode, _, err := client.Catalog().Node(name, nil)
		if err != nil {
			_ = c.Error(err)
			return
		}

		if catalogNode == nil {
			_ = c.Error(NotFoundError("could not find host"))
			return
		}

		host := hosts.NewHost(*catalogNode.Node, client)
		//c.HTML(http.StatusOK, "ha_checks.html.tmpl", gin.H{
		c.HTML(http.StatusOK, "ha_checks_tmp.html.tmpl", gin.H{
			"Hostname": host.Name(),
			//"HAChecks": host.HAChecks(),
			"HAChecks": r,
		})
	}
}
