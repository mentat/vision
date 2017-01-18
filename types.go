package vision

import (
	"fmt"
	"time"
)

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

func (e CheckEvent) String() string {
	return fmt.Sprintf("Node: %s, CheckID: %s, Name: %s, Status: %s, Notes: %s, Service:%s",
		e.Node, e.CheckID, e.Name, e.Status, e.Notes, e.ServiceID)
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

func (e ServiceEvent) String() string {
	return fmt.Sprintf("ID: %s, Name: %s, Tags: %s, Port: %d, Address: %s, Node:%+v",
		e.ID, e.Name, e.Tags, e.Port, e.Address, e.Node)
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
