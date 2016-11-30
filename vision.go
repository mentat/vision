package vision

import (
	"time"

	"github.com/hashicorp/consul/api"
)

// Service - a service on a VisionNode (like nginx, mysql, etc.)
type Service struct {
	ID            string
	NodeID        string
	Health        string
	Address       string
	Port          int
	CreatedAt     *time.Time
	Configuration map[string]interface{}
	Tags          []string
}

// Node - a node in an infra.
type Node struct {
	ID        string
	Zone      string
	Health    string
	Address   string
	CreatedAt *time.Time
	Services  []*Service
}

type consulServer struct {
	ID        string
	CreatedAt *time.Time
	Address   string
}

// Infra -
type Infra struct {
	Region       string
	consulClient *api.Client
}

// NewInfra -
func NewInfra(region string) (*Infra, error) {

	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return nil, err
	}

	status := client.Status()
	_, err = status.Leader()

	if err != nil {
		return nil, err
	}

	infra := &Infra{
		Region:       region,
		consulClient: client,
	}

	return infra, nil
}

// getConsulServers -
func (infra Infra) getConsulServers() ([]*consulServer, error) {
	return nil, nil
}

// GetNode -
func (infra Infra) GetNode(nodeID string) (*Node, error) {
	return nil, nil
}

// SetValue -
func (infra Infra) SetValue(key, value string) error {
	// Get a handle to the KV API
	kv := infra.consulClient.KV()

	// PUT a new KV pair
	p := &api.KVPair{Key: key, Value: []byte(value)}
	_, err := kv.Put(p, nil)
	if err != nil {
		return err
	}
	return nil
}

// GetValue -
func (infra Infra) GetValue(key string) (string, error) {

	kv := infra.consulClient.KV()
	pair, _, err := kv.Get(key, nil)
	if err != nil {
		return "", err
	}

	if len(pair.Value) == 0 {
		return "", nil
	}

	return string(pair.Value), nil
}

// GetServiceByName - get all the services by name, like "mongo" or "nginx".
func (infra Infra) GetServiceByName(name string) ([]*Service, error) {

	return infra.GetServiceByTag(name, "")
}

// GetServiceByTag - get all the services identified by name and tag, like "mongo", "master"
func (infra Infra) GetServiceByTag(name, tag string) ([]*Service, error) {

	cat := infra.consulClient.Catalog()
	services, _, err := cat.Service(name, tag, nil)
	if err != nil {
		return nil, err
	}

	retVal := make([]*Service, len(services))
	for _, s := range services {
		service := &Service{
			ID:            s.ServiceID,
			NodeID:        s.Node,
			Health:        "",
			Address:       s.Address,
			Port:          s.ServicePort,
			CreatedAt:     nil,
			Configuration: nil,
			Tags:          s.ServiceTags,
		}
		retVal = append(retVal, service)
	}
	return retVal, nil
}

// GetServiceByHealth -
func (infra Infra) GetServiceByHealth(health string) ([]*Service, error) {
	return nil, nil
}

// UpdateConfiguration -
func (infra Infra) UpdateConfiguration(serviceID string, config map[string]interface{}) error {
	return nil
}

// UpdateConfigurationByTag -
func (infra Infra) UpdateConfigurationByTag(serviceID string, config map[string]interface{}) error {
	return nil
}

// RunCommand - Run a command asyncronously on a node.
func (infra Infra) RunCommand(nodeID string, payload map[string]interface{}) error {
	return nil
}
