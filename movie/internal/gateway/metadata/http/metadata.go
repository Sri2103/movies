package http

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"

	"movieexample.com/metadata/pkg/model"
	"movieexample.com/movie/internal/gateway"
	"movieexample.com/pkg/discovery"
)

// Gateway represents a gateway for accessing metadata.
type Gateway struct {
	// MetadataURL is the URL of the metadata service.
	registry discovery.Registry
}

// New creates a new Gateway with the given metadataURL.
func New(registry discovery.Registry) *Gateway {
	return &Gateway{
		registry,
	}
}

// Get retrieves the metadata for the given ID from the metadata service.
// The context can be used to cancel or timeout the request.
// It returns the retrieved metadata or an error if the request failed.
func (g *Gateway) Get(ctx context.Context, id string) (*model.Metadata, error) {
	addrs, err := g.registry.ServiceAddresses(ctx, "metadata")
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("http://%s/%s", addrs[rand.Intn(len(addrs))], "metadata")

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	values := req.URL.Query()
	values.Add("id", id)
	req.URL.RawQuery = values.Encode()
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, gateway.ErrNotFound
	} else if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}
	var metadata model.Metadata
	if err := json.NewDecoder(res.Body).Decode(&metadata); err != nil {
		return nil, err
	}
	return &metadata, nil
}
