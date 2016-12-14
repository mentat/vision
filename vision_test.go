package vision

import (
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/mentat/cloudlayer"
)

func runContainer(name string) (cloudlayer.CloudLayer, string, error) {
	layer, _ := cloudlayer.NewCloudLayer("docker")
	inst, err := layer.CreateInstance(cloudlayer.InstanceDetails{
		BaseImage: "consul",
		ExposedPorts: []cloudlayer.PortDetails{
			cloudlayer.PortDetails{
				InstancePort: 8500,
				HostPort:     8500,
				Protocol:     "tcp",
			},
		},
	})

	if err != nil {
		return nil, "", err
	}

	for {
		if _, err := NewInfra("local"); err != nil {
			time.Sleep(50 * time.Millisecond)
		} else {
			break
		}
	}

	return layer, inst.ID, nil
}

func stopContainer(layer cloudlayer.CloudLayer, id string) {
	_, err := layer.RemoveInstance(id)
	if err != nil {
		fmt.Printf("Error stopping docker container: %s", err)
	}
}

func TestMain(m *testing.M) {
	flag.Parse()
	layer, id, err := runContainer("consul")

	if err != nil {
		fmt.Printf("Cannot start consul service: %s", err.Error())
		os.Exit(1)
	} else {
		retval := m.Run()
		stopContainer(layer, id)
		os.Exit(retval)
	}
}

func TestKV(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode.")
	}

	infra, err := NewInfra("local")
	if err != nil {
		t.Fatalf("Could not connect to Consul: %s", err)
	}

	err = infra.SetValue("foo", "bar")
	if err != nil {
		t.Fatalf("Could not set value to Consul: %s", err)
	}

	v, err := infra.GetValue("foo")
	if err != nil {
		t.Fatalf("Could not get value from Consul: %s", err)
	}

	if v != "bar" {
		t.Fatalf("Consul value is invalid.")
	}
}

func TestServices(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	infra, err := NewInfra("local")
	if err != nil {
		t.Fatalf("Could not connect to Consul.")
	}

	services, err := infra.GetServiceByName("consul")
	if err != nil {
		t.Fatalf("Could not list services.")
	}
	if len(services) == 0 {
		t.Fatalf("No services found in Consul.")
	}

}

func TestNodes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	infra, err := NewInfra("local")
	if err != nil {
		t.Fatalf("Could not connect to Consul.")
	}

	services, err := infra.GetServiceByName("consul")
	if err != nil {
		t.Fatalf("Could not list services.")
	}

	if len(services) == 0 {
		t.Fatalf("No services found in Consul.")
	}

	fmt.Printf("%v\n", services)
	node, err := infra.GetNode(services[0].NodeID)
	if err != nil {
		t.Fatalf("Could not get node: %s", services[0].NodeID)
	}

	fmt.Printf("%v\n", node)

}

func TestWatchPrefix(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	infra, err := NewInfra("local")
	if err != nil {
		t.Fatalf("Could not connect to Consul.")
	}
	done := make(chan bool, 1)

	events := infra.WatchKeys("foo/", done)

	infra.SetValue("foo/bar", "bar1")
	time.Sleep(20 * time.Millisecond)
	if event := <-events; event.KV.Value != "bar1" {
		t.Fatalf("Event equals: %s", event.KV.Value)
	}
	infra.SetValue("foo/bat", "bar2")
	time.Sleep(20 * time.Millisecond)
	if event := <-events; event.KV.Value != "bar2" {
		t.Fatalf("Event equals: %s", event.KV.Value)
	}
	infra.SetValue("foo/bas", "bar3")
	time.Sleep(20 * time.Millisecond)
	if event := <-events; event.KV.Value != "bar3" {
		t.Fatalf("Event equals: %s", event.KV.Value)
	}

	done <- true
}

func TestWatchEvent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	infra, err := NewInfra("local")
	if err != nil {
		t.Fatalf("Could not connect to Consul.")
	}
	done := make(chan bool, 1)

	events := infra.WatchEvent("node-456-logs", done)

	infra.FireEvent(&UserEvent{Name: "node-456-logs", Payload: []byte("1")})
	time.Sleep(20 * time.Millisecond)
	if event := <-events; string(event.UserEvent.Payload) != "1" {
		t.Fatalf("Event equals: %s", event.KV.Value)
	}

	infra.FireEvent(&UserEvent{Name: "node-456-logs", Payload: []byte("2")})
	time.Sleep(20 * time.Millisecond)
	if event := <-events; string(event.UserEvent.Payload) != "2" {
		t.Fatalf("Event equals: %s", event.KV.Value)
	}

	infra.FireEvent(&UserEvent{Name: "node-456-logs", Payload: []byte("3")})
	time.Sleep(20 * time.Millisecond)
	if event := <-events; string(event.UserEvent.Payload) != "3" {
		t.Fatalf("Event equals: %s", event.KV.Value)
	}

	done <- true
}

// Not currently working for some reason...
func testCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	infra, err := NewInfra("local")
	if err != nil {
		t.Fatalf("Could not connect to Consul.")
	}
	done := make(chan bool, 1)

	checks := infra.WatchChecks("mongo", done)

	infra.testRegisterService("foobar", "1.1.1.1", "foobar")
	infra.testDeregisterService("foobar", "1.1.1.1")

	if check := <-checks; string(check.Check.Status) != "3" {
		t.Fatalf("Check status equals: %s", check.Check.Status)
	}

	done <- true
}
