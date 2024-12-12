package discovery

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/exp/rand"
)

// Registry is an interface that provides methods for registering, deregistering, and
// looking up service instances in a discovery registry.
type Registry interface {
	// Register registers a new service instance with the discovery registry.
	// The instanceID parameter uniquely identifies the service instance.
	// The serviceName parameter specifies the name of the service.
	// The hostPort parameter specifies the host and port of the service instance.
	// The function returns an error if the registration fails.
	Register(ctx context.Context, instanceID string, serviceName string, hostPort string) error

	// DeRegister deregisters a service instance from the discovery registry.
	// The instanceID parameter uniquely identifies the service instance.
	// The serviceName parameter specifies the name of the service.
	// The function returns an error if the deregistration fails.
	DeRegister(ctx context.Context, instanceID string, serviceName string) error

	// ServiceAddresses returns the list of host:port addresses for the given service ID.
	// The ctx parameter is the context for the operation.
	// The serviceID parameter specifies the ID of the service to look up.
	// The function returns a slice of strings representing the host:port addresses for the service instances,
	// and an error if the lookup fails.
	ServiceAddresses(ctx context.Context, serviceID string) ([]string, error)

	// ReportHealthState reports the health state of a service instance to the discovery registry.
	// The instanceID parameter uniquely identifies the service instance.
	// The serviceName parameter specifies the name of the service.
	// The function returns an error if the health state reporting fails.
	ReportHealthState(instanceID string) error
}

// ErrNotFound is returned when no service addresses are found.
var ErrNotFound = errors.New("no service addresses found")

// GenerateInstanceID generates a unique instance ID for a service by combining the
// provided service name with a random integer.
func GenerateInstanceID(serviceName string) string {
	return fmt.Sprintf("%s-%d", serviceName, rand.New(rand.NewSource(uint64(time.Now().UnixNano()))).Int())
}
