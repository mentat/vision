# Vision

Infrastructure Management with Consul.

## Overview

All managed nodes in the infrastructure run Consul agent and consul template
(depending on workload) to manage discoverability, health checks, and
daemon configuration.


## Theoretical Workflow

    import "github.com/mentat/vision"
    infra, err := vision.NewVisionInfra("us-east-1", "152.1.1.206")

    // Look at some services...
    services, err := infra.GetServiceByTag("nginx")


    // Configure one of the services
    err := infra.UpdateConfiguration(
        service[0].ID,
        map[string]interface{}{
            "http.bind.port": 80
        })

    // Configure all of the services
    err := infra.UpdateConfigurationByTag(
        "nginx",
        map[string]interface{}{
            "http.bind.port": 80
        })

    // Run a command on a nodeId
    err := infra.RunCommand(service[0].NodeID, "backup", map[string]interface{}{
        "backup_id":"XYZ123"
    })
