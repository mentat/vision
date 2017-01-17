package vision

import (
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/consul/structs"
	"github.com/hashicorp/consul/watch"
	"github.com/juju/loggo"
)

var logger = loggo.GetLogger("vision")

// KeyPair - a key/value pair.
type KeyPair struct {
	Key   string
	Value string
}

// Service - a service on a VisionNode (like nginx, mysql, etc.)
type Service struct {
	ID              string
	Name            string
	NodeID          string
	Health          string
	Address         string
	TaggedAddresses map[string]string
	Port            int
	CreatedAt       *time.Time
	Configuration   map[string]interface{}
	Tags            []string
}

// Node - a node in an infra.
type Node struct {
	Node            string
	Zone            string
	Health          string
	Address         string
	TaggedAddresses map[string]string
	CreatedAt       *time.Time
	Services        []*Service
}

// NodeEvent - an event related to a node.
type NodeEvent struct {
	ID              string
	Address         string
	TaggedAddresses map[string]string
	Meta            map[string]string
}

func (n NodeEvent) String() string {
	return fmt.Sprintf("ID:%s Address:%s", n.ID, n.Address)
}

// CheckEvent - a health check event.
type CheckEvent struct {
	Node        string
	CheckID     string
	Name        string
	Status      string
	Notes       string
	Output      string
	ServiceID   string
	ServiceName string
}

// ServiceEvent - A Service Event.
type ServiceEvent struct {
	ID                string
	Name              string
	Tags              []string
	Port              int
	Address           string
	EnableTagOverride bool
	Node              NodeEvent
}

// UserEvent - A custom event type.
type UserEvent struct {
	ID            string
	Name          string
	Payload       []byte
	NodeFilter    string
	ServiceFilter string
	TagFilter     string
	Version       int
	LTime         uint64
}

// Event - an asyncronous event received via a watch.
type Event struct {
	Type string

	Check     *CheckEvent
	Node      *NodeEvent
	Service   *ServiceEvent
	UserEvent *UserEvent
	KV        *KeyPair
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
	logger.SetLogLevel(loggo.TRACE)

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

// GetNodesOfService -
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
		Node:            nodeID,
		Address:         node.Node.Address,
		TaggedAddresses: node.Node.TaggedAddresses,
		Services:        services,
	}

	return ret, nil
}

func (infra Infra) testRegisterService(name, address, nodeID string) error {

	catalog := infra.consulClient.Catalog()

	reg := &api.CatalogRegistration{
		Node:       name,
		Address:    address,
		Datacenter: infra.Region,
		Service: &api.AgentService{
			ID:      name,
			Service: name,
		},
		Check: &api.AgentCheck{
			Node:      nodeID,
			CheckID:   name,
			Name:      name,
			Status:    structs.HealthPassing,
			ServiceID: name,
		},
	}
	_, err := catalog.Register(reg, nil)
	if err != nil {
		return err
	}
	return nil
}

func (infra Infra) testDeregisterService(name, address string) error {
	catalog := infra.consulClient.Catalog()

	dereg := &api.CatalogDeregistration{
		Node:       name,
		Address:    address,
		Datacenter: infra.Region,
	}

	catalog.Deregister(dereg, nil)
	return nil
}

// FireEvent - fire a custom event into the cluster.
func (infra Infra) FireEvent(event *UserEvent) error {
	e := infra.consulClient.Event()
	params := &api.UserEvent{
		Name:    event.Name,
		Payload: event.Payload,
	}

	if _, _, err := e.Fire(params, nil); err != nil {
		return err
	}

	return nil
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
			ID:              s.ServiceID,
			NodeID:          s.Node,
			Health:          "",
			Address:         s.Address,
			TaggedAddresses: s.TaggedAddresses,
			Name:            s.ServiceName,
			Port:            s.ServicePort,
			CreatedAt:       nil,
			Configuration:   nil,
			Tags:            s.ServiceTags,
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
			Type: kind,
			Check: &CheckEvent{
				Status:      val.Status,
				ServiceName: val.ServiceName,
				ServiceID:   val.ServiceID,
				CheckID:     val.CheckID,
				Node:        val.Node,
				Output:      val.Output,
				Notes:       val.Notes,
			},
		}
		out <- event
	}
}

func (infra Infra) decodeKeyPairs(kind string, pairs *api.KVPairs, out chan Event, idx uint64) {

	for _, val := range *pairs {

		if idx > val.ModifyIndex {
			continue
		}

		event := Event{
			Type: kind,
			KV: &KeyPair{
				Key:   val.Key,
				Value: string(val.Value),
			},
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
			Type: kind,
			KV: &KeyPair{
				Key:   val.Key,
				Value: string(val.Value),
			},
		}
		out <- event
	}
}

func (infra Infra) decodeService(kind string, services []*api.ServiceEntry, out chan Event) {
	for _, val := range services {

		event := Event{
			Type: kind,
			Service: &ServiceEvent{
				Tags: val.Service.Tags,
				Name: val.Service.Service,
				ID:   val.Service.ID,
				Node: NodeEvent{
					ID:              val.Node.Node,
					Address:         val.Node.Address,
					TaggedAddresses: val.Node.TaggedAddresses,
				},
				Address:           val.Service.Address,
				Port:              val.Service.Port,
				EnableTagOverride: val.Service.EnableTagOverride,
			},
		}
		out <- event
	}
}

func (infra Infra) decodeNode(kind string, nodes []*api.Node, out chan Event) {
	for _, val := range nodes {

		event := Event{
			Type: kind,
			Node: &NodeEvent{
				ID:              val.Node,
				Address:         val.Address,
				TaggedAddresses: val.TaggedAddresses,
			},
		}
		out <- event
	}
}

func (infra Infra) decodeEvent(kind string, nodes []*api.UserEvent, out chan Event) {
	for _, val := range nodes {

		event := Event{
			Type: kind,
			UserEvent: &UserEvent{
				ID:            val.ID,
				Name:          val.Name,
				Payload:       val.Payload,
				NodeFilter:    val.NodeFilter,
				ServiceFilter: val.ServiceFilter,
				TagFilter:     val.TagFilter,
				Version:       val.Version,
				LTime:         val.LTime,
			},
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

		logger.Infof("Handler...%s, Index: %d", kind, idx)

		switch kind {
		case "checks":
			if v, ok := raw.([]*api.HealthCheck); ok && len(v) > 0 {
				infra.decodeCheck("check", v, out, idx)
			} else {
				logger.Errorf("Error in check watch: %#v", raw)
			}
		case "keyprefix":
			if v, ok := raw.(api.KVPairs); ok && len(v) > 0 {
				infra.decodeKeyPairs("keyprefix", &v, out, idx)
			} else {
				logger.Errorf("Error in keyprefix watch: %#v", raw)
			}
		case "services":
			if v, ok := raw.([]*api.ServiceEntry); ok && len(v) > 0 {
				infra.decodeService("service", v, out)
			} else {
				logger.Errorf("Error in service watch: %#v", raw)
			}
		case "nodes":
			if v, ok := raw.([]*api.Node); ok && len(v) > 0 {
				infra.decodeNode("nodes", v, out)
			} else {
				logger.Errorf("Error in node watch: %#v", raw)
			}
		case "key":
			if v, ok := raw.([]*api.KVPair); ok && len(v) > 0 {
				infra.decodeKey("key", v, out, idx)
			} else {
				logger.Errorf("Error in key watch: %#v", raw)
			}
		case "event":
			if v, ok := raw.([]*api.UserEvent); ok && len(v) > 0 {
				infra.decodeEvent("event", v, out)
			} else {
				logger.Errorf("Error in event watch: %#v", raw)
			}
		}
	}

	go func() {
		defer close(out)

		for {
			select {
			case <-done:
				logger.Warningf("Stopping plan...")
				plan.Stop()
				return
			default:
				logger.Infof("Running plan...")
				plan.Run(infra.consulConfig.Address)
			}
		}
	}()
	return out

}

// WatchChecks - Watch the health checks for a particular service.
func (infra Infra) WatchChecks(service string, done <-chan bool) <-chan Event {
	return infra.doWatch(map[string]interface{}{
		"type":    "checks",
		"service": service,
	}, done)
}

// WatchCheckState - Watch the health checks for a state: ie, any, passing, warning, or critical
func (infra Infra) WatchCheckState(state string, done <-chan bool) <-chan Event {
	return infra.doWatch(map[string]interface{}{
		"type":  "checks",
		"state": state,
	}, done)
}

// WatchKeys - Watch the key prefixes given.
func (infra Infra) WatchKeys(prefix string, done <-chan bool) <-chan Event {
	return infra.doWatch(map[string]interface{}{
		"type":   "keyprefix",
		"prefix": prefix,
	}, done)
}

// WatchService - Watch a particular service.
func (infra Infra) WatchService(service string, done <-chan bool) <-chan Event {
	return infra.doWatch(map[string]interface{}{
		"type":    "service",
		"service": service,
	}, done)
}

// WatchServices - Watch all services.
func (infra Infra) WatchServices(done <-chan bool) <-chan Event {
	return infra.doWatch(map[string]interface{}{
		"type": "services",
	}, done)
}

// WatchNodes - Watch all services.
func (infra Infra) WatchNodes(done <-chan bool) <-chan Event {
	return infra.doWatch(map[string]interface{}{
		"type": "nodes",
	}, done)
}

// WatchEvent - Watch all events of a certain type..
func (infra Infra) WatchEvent(name string, done <-chan bool) <-chan Event {
	return infra.doWatch(map[string]interface{}{
		"type": "event",
		"name": name,
	}, done)
}
