package vision

import "time"

// VisionService - a service on a VisionNode (like nginx, mysql, etc.)
type VisionService struct {
	ID            string
	NodeID        string
	Health        string
	Port          int
	CreatedAt     *time.Time
	Configuration map[string]interface{}
	Tags          []string
}

// VisionNode - a node in an infra.
type VisionNode struct {
	ID        string
	Host      string
	Zone      string
	Health    string
	Address   string
	CreatedAt *time.Time
}

type ConsulServer struct {
	ID        string
	CreatedAt *time.Time
	Address   string
}

type VisionInfra struct {
	Region string
}

func NewVisionInfra(region, serverAddress string) (*VisionInfra, error) {
	return nil, nil
}

func (this VisionInfra) GetConsulServers() ([]*ConsulServer, error) {
	return nil, nil
}

func (this VisionInfra) GetNode(nodeId string) (*VisionNode, error) {
	return nil, nil
}

func (this VisionInfra) GetService(serviceId string) (*VisionService, error) {
	return nil, nil
}

func (this VisionInfra) GetServiceByTag(tag string) ([]*VisionService, error) {
	return nil, nil
}

func (this VisionInfra) GetServiceByHealth(health string) ([]*VisionService, error) {
	return nil, nil
}

func (this VisionInfra) UpdateConfiguration(serviceId string, config map[string]interface{}) error {
	return nil
}

func (this VisionInfra) UpdateConfigurationByTag(serviceId string, config map[string]interface{}) error {
	return nil
}

// Run a command asyncronously on a node.
func (this VisionInfra) RunCommand(nodeId string, payload map[string]interface{}) error {
	return nil
}
