package store

import (
	"fmt"

	osb "github.com/pmorie/go-open-service-broker-client/v2"
)

type ServiceInstance struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ServiceID   string `json:"service_id"`
	ServiceName string `json:"service_name"`
	PlanID      string `json:"plan_id"`
	PlanName    string `json:"plan_name"`
}

type Store struct {
	ServiceInstances map[string]ServiceInstance
	Catalog          osb.CatalogResponse
}

func NewStore(catalog osb.CatalogResponse) *Store {
	serviceInstances := make(map[string]ServiceInstance)

	return &Store{
		ServiceInstances: serviceInstances,
		Catalog:          catalog,
	}
}

func (s *Store) ListServiceInstances() []ServiceInstance {
	instanceList := []ServiceInstance{}
	for _, v := range s.ServiceInstances {
		instanceList = append(instanceList, v)
	}

	return instanceList
}

func (s *Store) CreateServiceInstance(name, serviceID, planID string) error {
	serviceInst := ServiceInstance{
		ID:          name,
		Name:        name,
		ServiceID:   serviceID,
		ServiceName: s.getServiceNameForID(serviceID),
		PlanID:      planID,
		PlanName:    s.getPlanNameForID(planID),
	}

	if _, ok := s.ServiceInstances[name]; ok {
		return fmt.Errorf("service instance %s already exists", name)
	}

	s.ServiceInstances[name] = serviceInst
	return nil
}

func (s *Store) GetCatalog() osb.CatalogResponse {
	return s.Catalog
}

func (s *Store) getServiceNameForID(id string) string {
	for _, v := range s.Catalog.Services {
		if v.ID == id {
			return v.Name
		}
	}
	return ""
}

func (s *Store) getPlanNameForID(planID string) string {
	for _, service := range s.Catalog.Services {
		for _, plan := range service.Plans {
			if plan.ID == planID {
				return plan.Name
			}
		}
	}
	return ""
}
