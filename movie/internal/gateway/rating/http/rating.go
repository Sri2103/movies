package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"k8s.io/apimachinery/pkg/util/rand"
	"movieexample.com/movie/internal/gateway"
	"movieexample.com/pkg/discovery"
	"movieexample.com/rating/pkg/model"
)

// Gateway is a struct that holds a discovery Registry.
// It is used to interact with the rating service.
type Gateway struct {
	registry discovery.Registry
}

// New creates a new instance of the Gateway struct with the provided discovery Registry.
func New(registry discovery.Registry) *Gateway {
	return &Gateway{
		registry,
	}
}

// GetAggregatedRating retrieves the aggregated rating for the specified record ID and record type.
// The context parameter is used to control the lifetime of the request.
// The recordID parameter specifies the unique identifier of the record.
// The recordType parameter specifies the type of the record.
// The function returns the aggregated rating as a float64 value, and an error if any occurred during the operation.
func (g *Gateway) GetAggregatedRating(ctx context.Context, recordID model.RecordID, recordType model.RecordType) (float64, error) {
	// req, err := http.NewRequest("GET", g.ratingURL+"/rating", nil)
	addrs, err := g.registry.ServiceAddresses(ctx, "metadata")
	if err != nil {
		return 0, err
	}
	url := fmt.Sprintf("http://%s/%s", addrs[rand.Intn(len(addrs))], "rating")

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return 0, err
	}
	req = req.WithContext(ctx)
	values := req.URL.Query()
	values.Add("id", string(recordID))
	values.Add("type", string(recordType))
	req.URL.RawQuery = values.Encode()
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return 0, gateway.ErrNotFound
	} else if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("non 200 response: %v", res)
	}
	var rating float64
	if err := json.NewDecoder(res.Body).Decode(&rating); err != nil {
		return 0, err
	}
	return rating, nil
}

// PutRating updates the rating for the specified record ID and record type.
// The context parameter is used to control the lifetime of the request.
// The recordID parameter specifies the unique identifier of the record.
// The recordType parameter specifies the type of the record.
// The rating parameter specifies the new rating to be set.
// The function returns an error if any occurred during the operation.
func (g *Gateway) PutRating(ctx context.Context, recordID model.RecordID, recordType model.RecordType, rating *model.Rating) error {
	addrs, err := g.registry.ServiceAddresses(ctx, "rating")
	if err != nil {
		return err
	}
	url := fmt.Sprintf("http://%s/%s", addrs[rand.Intn(len(addrs))], "rating")
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	values := req.URL.Query()
	values.Add("id", string(recordID))
	values.Add("type", string(recordType))
	values.Add("value", strconv.Itoa(int(rating.Value)))
	values.Add("userId", string(rating.UserID))
	req.URL.RawQuery = values.Encode()
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return gateway.ErrNotFound
	} else if res.StatusCode != http.StatusOK {
		return fmt.Errorf("non 200 response: %v", res)
	}
	return nil
}
