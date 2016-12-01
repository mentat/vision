package vision

import (
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/watch"
	"github.com/juju/loggo"
)

var logger = loggo.GetLogger("vision")

// Service - a service on a VisionNode (like nginx, mysql, etc.)
type Service struct {
	ID            string
	Name          string
	NodeID        string
	Health        string
	Address       string
	TaggedAddress map[string]string
	Port          int
	CreatedAt     *time.Time
	Configuration map[string]interface{}
	Tags          []string
}

// Node - a node in an infra.
type Node struct {
	ID            string
	Zone          string
	Health        string
	Address       string
	TaggedAddress map[string]string
	CreatedAt     *time.Time
	Services      []*Service
}

// Event - an asyncronous event received via a watch.
type Event struct {
	Type          string
	ServiceName   string
	ServiceID     string
	Address       string
	Tags          []string
	Port          int
	CheckID       string
	Node          string
	Message       string
	Status        string
	Key           string
	Value         string
	TaggedAddress map[string]string
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
	consulConfig *api.Config
}

// NewInfra -
func NewInfra(region string) (*Infra, error) {

	config := api.DefaultConfig()
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	status := client.Status()
	leader, err := status.Leader()

	if err != nil {
		return nil, err
	}

	if leader == "" {
		return nil, fmt.Errorf("Invalid leader.")
	}

	infra := &Infra{
		Region:       region,
		consulClient: client,
		consulConfig: config,
	}

	return infra, nil
}

// getConsulServers -
func (infra Infra) getConsulServers() ([]*consulServer, error) {
	return nil, nil
}

// GetNodes -
func (infra Infra) GetNodesOfService(name string) ([]*Node, error) {

	return nil, nil
}

// GetNode -
func (infra Infra) GetNode(nodeID string) (*Node, error) {

	cat := infra.consulClient.Catalog()
	node, _, err := cat.Node(nodeID, nil)

	if err != nil {
		return nil, err
	}

	if node == nil {
		logger.Errorf("Node is nil, clean me up.")
		return nil, fmt.Errorf("Node does not exist.")
	}

	services := make([]*Service, 0, len(node.Services))
	for _, v := range node.Services {
		service := &Service{
			ID:      v.ID,
			NodeID:  nodeID,
			Tags:    v.Tags,
			Port:    v.Port,
			Address: v.Address,
			Name:    v.Service,
		}
		services = append(services, service)
	}

	ret := &Node{
		ID:            nodeID,
		Address:       node.Node.Address,
		TaggedAddress: node.Node.TaggedAddresses,
		Services:      services,
	}

	return ret, nil
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

	if pair == nil || len(pair.Value) == 0 {
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
	for i, s := range services {
		service := &Service{
			ID:            s.ServiceID,
			NodeID:        s.Node,
			Health:        "",
			Address:       s.Address,
			TaggedAddress: s.TaggedAddresses,
			Name:          s.ServiceName,
			Port:          s.ServicePort,
			CreatedAt:     nil,
			Configuration: nil,
			Tags:          s.ServiceTags,
		}
		retVal[i] = service
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

func (infra Infra) decodeCheck(kind string, checks []*api.HealthCheck, out chan Event, idx uint64) {
	for _, val := range checks {

		event := Event{
			Type:        kind,
			Status:      val.Status,
			ServiceName: val.ServiceName,
			ServiceID:   val.ServiceID,
			CheckID:     val.CheckID,
			Node:        val.Node,
			Message:     val.Output,
		}
		out <- event
	}
}

func (infra Infra) decodeKeyPairs(kind string, pairs *api.KVPairs, out chan Event, idx uint64) {
	logger.Errorf("Key Pair Length: %d", len(*pairs))
	for _, val := range *pairs {

		if idx > val.ModifyIndex {
			continue
		}

		event := Event{
			Type:  kind,
			Key:   val.Key,
			Value: string(val.Value),
		}
		out <- event

	}
}

func (infra Infra) decodeKey(kind string, keys []*api.KVPair, out chan Event, idx uint64) {
	for _, val := range keys {

		if idx > val.ModifyIndex {
			continue
		}

		event := Event{
			Type:  kind,
			Key:   val.Key,
			Value: string(val.Value),
		}
		out <- event
	}
}

func (infra Infra) decodeService(kind string, services []*api.ServiceEntry, out chan Event) {
	for _, val := range services {

		event := Event{
			Type:        kind,
			Tags:        val.Service.Tags,
			ServiceName: val.Service.Service,
			ServiceID:   val.Service.ID,
			Node:        val.Node.Node,
			Address:     val.Service.Address,
			Port:        val.Service.Port,
		}
		out <- event
	}
}

func (infra Infra) decodeNode(kind string, nodes []*api.Node, out chan Event) {
	for _, val := range nodes {

		event := Event{
			Type:          kind,
			Address:       val.Address,
			Node:          val.Node,
			TaggedAddress: val.TaggedAddresses,
		}
		out <- event
	}
}

func (infra Infra) doWatch(details map[string]interface{}, done <-chan bool) <-chan Event {
	out := make(chan Event, 10)

	kind := details["type"].(string)
	plan, _ := watch.Parse(details)

	plan.Handler = func(idx uint64, raw interface{}) {

		if raw == nil {
			return
		}

		logger.Errorf("Handler...%s, Index: %d", kind, idx)

		switch kind {
		case "check":
			if v, ok := raw.([]*api.HealthCheck); ok && len(v) > 0 {
				infra.decodeCheck("check", v, out, idx)
			} else {
				logger.Errorf("Error in check watch: %#v", raw)
			}
		case "keyprefix":
			if v, ok := raw.(api.KVPairs); ok && len(v) > 0 {
				logger.Errorf("Got Key Pairs")
				infra.decodeKeyPairs("keyprefix", &v, out, idx)
			} else {
				logger.Errorf("Error in check watch: %#v", raw)
			}
		case "service":
			if v, ok := raw.([]*api.ServiceEntry); ok && len(v) > 0 {
				infra.decodeService("service", v, out)
			} else {
				logger.Errorf("Error in check watch: %#v", raw)
			}
		case "nodes":
			if v, ok := raw.([]*api.Node); ok && len(v) > 0 {
				infra.decodeNode("nodes", v, out)
			} else {
				logger.Errorf("Error in check watch: %#v", raw)
			}
		case "key":
			if v, ok := raw.([]*api.KVPair); ok && len(v) > 0 {
				infra.decodeKey("key", v, out, idx)
			} else {
				logger.Errorf("Error in check watch: %#v", raw)
			}
		}
	}

	go func() {
		defer close(out)

		for {
			select {
			case <-done:
				logger.Errorf("Stopping...")
				plan.Stop()
				return
			default:
				logger.Errorf("Running...")
				plan.Run(infra.consulConfig.Address)
			}
		}
	}()
	return out

}

// WatchChecks - Watch the health checks for a particular service.
func (infra Infra) WatchChecks(service string, done <-chan bool) <-chan Event {
	return infra.doWatch(map[string]interface{}{
		"type":    "check",
		"service": service,
	}, done)
}

// WatchKeys - Watch the key prefixes given.
func (infra Infra) WatchKeys(prefix string, done <-chan bool) <-chan Event {
	return infra.doWatch(map[string]interface{}{
		"type":   "keyprefix",
		"prefix": prefix,
	}, done)
}
