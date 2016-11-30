package vision

import (
	"fmt"
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

	time.Sleep(500 * time.Millisecond)

	return layer, inst.ID, nil
}

func stopContainer(layer cloudlayer.CloudLayer, id string) {
	_, err := layer.RemoveInstance(id)
	if err != nil {
		fmt.Printf("Error stopping docker container: %s", err)
	}
}

func TestKV(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode.")
	}
	layer, id, err := runContainer("consul")

	if err != nil {
		t.Fatalf(err.Error())
	} else {
		defer stopContainer(layer, id)
	}

	infra, err := NewInfra("local")
	if err != nil {
		t.Fatalf("Could not connect to Consul.")
	}
	infra.SetValue("foo", "bar")
	v, err := infra.GetValue("foo")
	if err != nil {
		t.Fatalf("Could not get value from Consul.")
	}

	if v != "bar" {
		t.Fatalf("Consul value is invalid.")
	}
}

func _TestServices(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	infra, err := NewInfra("local")
	if err != nil {
		t.Fatalf("Could not connect to Consul.")
	}
	infra.SetValue("foo", "bar")
	v, err := infra.GetValue("foo")
	if err != nil {
		t.Fatalf("Could not get value from Consul.")
	}
	if v != "bar" {
		t.Fatalf("Consul value is invalid.")
	}
}

func _TestWatch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	infra, err := NewInfra("local")
	if err != nil {
		t.Fatalf("Could not connect to Consul.")
	}
	infra.SetValue("foo", "bar")
	v, err := infra.GetValue("foo")
	if err != nil {
		t.Fatalf("Could not get value from Consul.")
	}
	if v != "bar" {
		t.Fatalf("Consul value is invalid.")
	}
}
