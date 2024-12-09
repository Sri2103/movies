package consul

import (
	"context"
	"errors"
	"strconv"
	"strings"

	consul "github.com/hashicorp/consul/api"
)

// Registry is a Consul registry client that can be used to interact with a Consul
// cluster for service discovery and registration.
type Registry struct {
	client *consul.Client
}

// NewRegistry creates a new Consul registry client with the given address.
// It returns the registry client and any error that occurred during creation.
func NewRegistry(address string) (*Registry, error) {
	config := consul.DefaultConfig()
	config.Address = address
	client, err := consul.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &Registry{client: client}, nil
}

// Register registers a service with the Consul agent. It takes the context, an instance ID, a service name, and a host:port string.
// It returns an error if the host:port format is invalid or if there is an error registering the service with Consul.
func (r *Registry) Register(ctx context.Context, instanceID string, serviceName string, hostPort string) error {
	parts := strings.Split(hostPort, ":")
	if len(parts) != 2 {
		return errors.New("invalid host:port format, example: localhost:8081")
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return err
	}
	err = r.client.Agent().ServiceRegister(&consul.AgentServiceRegistration{
		ID:      instanceID,
		Name:    serviceName,
		Port:    port,
		Address: parts[0],
		Check:   &consul.AgentServiceCheck{CheckID: instanceID, TTL: "5s"},
	})
	if err != nil {
		return err
	}
	return nil
}

// DeRegister deregisters a service from the Consul agent. It takes the context, an instance ID, and a service name.
// It returns an error if there is an error deregistering the service with Consul.
func (r *Registry) DeRegister(ctx context.Context, instanceID string, serviceName string) error {
	return r.client.Agent().ServiceDeregister(instanceID)
}

// ServiceAddresses returns a list of service addresses for the given service ID.
// It queries the Consul agent for healthy service instances and returns their
// host:port addresses. If no service instances are found, it returns an error.
func (r *Registry) ServiceAddresses(ctx context.Context, serviceID string) ([]string, error) {
	services, _, err := r.client.Health().Service(serviceID, "", true, nil)
	if err != nil {
		return nil, err
	} else if len(services) == 0 {
		return nil, errors.New("no service instances found")
	}
	addresses := make([]string, len(services))
	for i, service := range services {
		addresses[i] = service.Service.Address + ":" + strconv.Itoa(service.Service.Port)
	}
	return addresses, nil
}

// ReportHealthyState reports the healthy state of a service instance to the Consul agent.
// It takes the instance ID and service name, and passes the TTL for the service's health check.
// This allows the Consul agent to mark the service instance as healthy.
func (r *Registry) ReportHealthState(instanceID string, serviceName string) error {
	return r.client.Agent().PassTTL(instanceID, "")
}
