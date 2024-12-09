package movie

import (
	"context"
	"errors"

	metadatamodel "movieexample.com/metadata/pkg/model"
	"movieexample.com/movie/internal/gateway"
	"movieexample.com/movie/pkg/model"
	ratingmodel "movieexample.com/rating/pkg/model"
)

var (
	ErrNotFound = errors.New("movie metadata not found")
)

// ratingGateway is an interface that provides methods for interacting with a rating system.
// GetAggregatedRating retrieves the aggregated rating for a given record ID and record type.
// PutRating stores a new rating for the given record ID and record type.
type ratingGateway interface {
	GetAggregatedRating(ctx context.Context, recordID ratingmodel.RecordID, recordType ratingmodel.RecordType) (float64, error)
	PutRating(ctx context.Context, recordID ratingmodel.RecordID, recordType ratingmodel.RecordType, rating *ratingmodel.Rating) error
}

// metadataGateway is an interface that provides methods for interacting with a metadata system.
// Get retrieves the metadata for the given ID.
type metadataGateway interface {
	Get(ctx context.Context, id string) (*metadatamodel.Metadata, error)
}

// Controller is the main struct for the movie controller. It contains the necessary gateways
// for interacting with the rating and metadata systems.
type Controller struct {
	ratingGateway   ratingGateway
	metadataGateway metadataGateway
}

// New creates a new instance of the Controller struct, which is the main struct for the movie controller.
// It takes two parameters: a ratingGateway and a metadataGateway, which are used to interact with the rating and metadata systems, respectively.
// The returned *Controller is ready to be used for handling movie-related operations.
func New(ratingGateway ratingGateway, metadataGateway metadataGateway) *Controller {
	return &Controller{
		ratingGateway:   ratingGateway,
		metadataGateway: metadataGateway,
	}
}

// Get retrieves the movie details for the given ID. It fetches the movie metadata from the
// metadataGateway and the aggregated rating from the ratingGateway, and returns a
// model.MovieDetails struct containing the metadata and rating.
// If the movie metadata or rating is not found, it returns ErrNotFound.
func (c *Controller) Get(ctx context.Context, id string) (*model.MovieDetails, error) {
	if id == "" {
		return nil, errors.New("empty id")
	}
	metadata, err := c.metadataGateway.Get(ctx, id)
	if err != nil && errors.Is(err, gateway.ErrNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}
	details := &model.MovieDetails{
		Metadata: *metadata,
	}
	rating, err := c.ratingGateway.GetAggregatedRating(ctx, ratingmodel.RecordID(id), ratingmodel.RecordTypeMovie)
	if err != nil && errors.Is(err, gateway.ErrNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}
	details.Rating = &rating
	return details, nil
}
