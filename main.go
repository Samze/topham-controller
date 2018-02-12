package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/pivotal-cf-experimental/topham-controller/api"
	"github.com/pivotal-cf-experimental/topham-controller/store"

	osb "github.com/pmorie/go-open-service-broker-client/v2"
)

var brokerURL = os.Getenv("BROKER_URL")
var username = os.Getenv("BROKER_USERNAME")
var password = os.Getenv("BROKER_PASSWORD")

var instancesStore *store.Store

func main() {
	r := mux.NewRouter()

	brokerClient := getBrokerClient()

	catalog, err := brokerClient.GetCatalog()
	if err != nil {
		log.Fatal(err)
	}

	instancesStore = store.NewStore(*catalog)

	ctrl := api.NewServicesController(brokerClient, instancesStore)

	r.HandleFunc("/v2/catalog", ctrl.CatalogHandler)
	r.HandleFunc("/v2/service_instances/{name}", ctrl.ProvisionHandler)
	r.HandleFunc("/v2/service_instances", ctrl.ListInstancesHandler)
	log.Fatal(http.ListenAndServe(":8080", r))
}

func getBrokerClient() osb.Client {
	config := osb.DefaultClientConfiguration()
	config.URL = brokerURL
	config.AuthConfig = &osb.AuthConfig{
		BasicAuthConfig: &osb.BasicAuthConfig{
			Username: username,
			Password: password,
		},
	}

	client, err := osb.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	return client
}
