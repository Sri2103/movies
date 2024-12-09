package memory

import (
	"context"
	"errors"
	"sync"
	"time"

	"movieexample.com/pkg/discovery"
)

// type serviceName string

// type instanceID string

type serviceInstance struct {
	hostPort   string
	lastActive time.Time
}

// Registry is a thread-safe in-memory registry for tracking service instances.
// It maps service names to a map of instance IDs to service instance details.
type Registry struct {
	sync.RWMutex
	serviceAddrs map[string]map[string]*serviceInstance
}

// NewRegistry creates a new thread-safe in-memory registry for tracking service instances.
// It initializes the serviceAddrs map to store service names mapped to instance IDs and service instance details.
func NewRegistry() *Registry {
	return &Registry{
		serviceAddrs: make(map[string]map[string]*serviceInstance),
	}
}

// Register registers a new service instance with the given instanceID, serviceName, and hostPort.
// It adds the service instance to the registry, creating a new entry for the serviceName if necessary.
// The lastActive field of the service instance is set to the current time.
// The function returns nil on success.
func (r *Registry) Register(ctx context.Context, instanceID string, serviceName string, hostPort string) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.serviceAddrs[serviceName]; !ok {
		r.serviceAddrs[serviceName] = make(map[string]*serviceInstance)
	}
	r.serviceAddrs[serviceName][instanceID] = &serviceInstance{
		hostPort:   hostPort,
		lastActive: time.Now(),
	}
	return nil
}

// DeRegister removes the service instance with the given instanceID and serviceName from the registry.
// If the service name no longer has any registered instances, the service name is also removed from the registry.
// The function returns nil on success.
func (r *Registry) DeRegister(ctx context.Context, instanceID string, serviceName string) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.serviceAddrs[serviceName]; ok {
		delete(r.serviceAddrs[serviceName], instanceID)
		if len(r.serviceAddrs[serviceName]) == 0 {
			delete(r.serviceAddrs, serviceName)
		}
	}
	return nil
}

// ServiceAddresses returns a list of active service instance addresses for the given serviceName.
// It retrieves the service instances from the registry, filters out any instances that have been inactive
// for more than 5 seconds, and returns the host:port addresses of the remaining active instances.
// If no service instances are found for the given serviceName, it returns discovery.ErrNotFound.
func (r *Registry) ServiceAddresses(ctx context.Context, serviceName string) ([]string, error) {
	r.RLock()
	defer r.RUnlock()
	serviceAddrs, ok := r.serviceAddrs[serviceName]
	if !ok {
		return nil, discovery.ErrNotFound
	}
	addresses := make([]string, 0, len(serviceAddrs))
	for _, instance := range serviceAddrs {
		if instance.lastActive.Before(time.Now().Add(-15 * time.Minute)) {
			continue
		}
		addresses = append(addresses, instance.hostPort)
	}
	return addresses, nil
}

func (r *Registry) ReportHealthState(instanceID string, serviceName string) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.serviceAddrs[serviceName]; !ok {
		return errors.New("service not registered yet")
	}

	if _, ok := r.serviceAddrs[serviceName]; ok {
		return errors.New("service not registered yet")
	}
	return nil
}
