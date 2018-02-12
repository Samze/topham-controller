package api

import (
	"encoding/json"
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

	details := brokerapi.ProvisionDetails{}
	err := json.NewDecoder(r.Body).Decode(&details)

	if err != nil {
		http.Error(w, "Error parsing provision request", http.StatusBadRequest)
		return
	}

	name := vars["name"]

	preq := osb.ProvisionRequest{
		InstanceID:        name,
		ServiceID:         details.ServiceID,
		PlanID:            details.PlanID,
		OrganizationGUID:  "dummy-org-id",
		SpaceGUID:         "dummy-space-id",
		AcceptsIncomplete: true,
	}

	provisionResponse, err := s.client.ProvisionInstance(&preq)

	if err != nil {
		http.Error(w, "Broker error:"+err.Error(), http.StatusBadGateway)
		return
	}

	resp, err := json.Marshal(provisionResponse)
	if err != nil {
		http.Error(w, "Broker error:"+err.Error(), http.StatusBadGateway)
		return
	}

	err = s.store.CreateServiceInstance(name, details.ServiceID, details.PlanID)
	if err != nil {
		http.Error(w, "Failed to save:"+err.Error(), http.StatusInternalServerError)
		return
	}

	if provisionResponse.Async {
		w.WriteHeader(202)
	}

	w.Write(resp)
}

func (s *ServicesController) CatalogHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := json.Marshal(s.store.GetCatalog())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(resp)
}

func (s *ServicesController) LastOperation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	req := osb.LastOperationRequest{
		InstanceID: vars["name"],
	}

	resp, err := s.client.PollLastOperation(&req)
	if err != nil {
		http.Error(w, "Broker error:"+err.Error(), http.StatusBadGateway)
		return
	}

	bytes, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Error parsing last operation response", http.StatusBadGateway)
		return
	}

	w.Write(bytes)
}

func (s *ServicesController) ListInstancesHandler(w http.ResponseWriter, r *http.Request) {
	instances := s.store.ListServiceInstances()
	bytes, err := json.Marshal(instances)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Write(bytes)
}
