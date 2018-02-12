package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pivotal-cf-experimental/topham-controller/store"
	"github.com/pivotal-cf/brokerapi"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
)

type Store interface {
	CreateServiceInstance(name, planID, serviceID string) error
	ListServiceInstances() []store.ServiceInstance
	GetCatalog() osb.CatalogResponse
}

type ServicesController struct {
	client osb.Client
	store  Store
}

func NewServicesController(client osb.Client, store Store) *ServicesController {
	return &ServicesController{
		client: client,
		store:  store,
	}
}

func (s *ServicesController) ProvisionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.Fatal(err)
	}

	details := brokerapi.ProvisionDetails{}
	err = json.Unmarshal(b, &details)
	if err != nil {
		log.Fatal(err)
	}

	name := vars["name"]

	preq := osb.ProvisionRequest{
		InstanceID:       name,
		ServiceID:        details.ServiceID,
		PlanID:           details.PlanID,
		OrganizationGUID: "dummy-org-id",
		SpaceGUID:        "dummy-space-id",
	}

	provisionResponse, err := s.client.ProvisionInstance(&preq)
	if err != nil {
		log.Fatal(err)
	}

	err = s.store.CreateServiceInstance(name, details.ServiceID, details.PlanID)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := json.Marshal(provisionResponse)
	if err != nil {
		log.Fatal(err)
	}

	w.Write(resp)
}

func (s *ServicesController) CatalogHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := json.Marshal(s.store.GetCatalog())
	if err != nil {
		log.Fatal(err)
	}

	w.Write(resp)
}

func (s *ServicesController) ListInstancesHandler(w http.ResponseWriter, r *http.Request) {
	instances := s.store.ListServiceInstances()
	bytes, err := json.Marshal(instances)
	if err != nil {
		log.Fatal(err)
	}

	w.Write(bytes)
}
