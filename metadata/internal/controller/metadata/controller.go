package metadata

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	"movieexample.com/metadata/internal/repository"
	"movieexample.com/metadata/pkg/model"
)

var ErrNotFound = errors.New("not found")

// metadataRepository defines the interface for interacting with the metadata repository.
// The Get method retrieves a metadata record by its ID.
type metadataRepository interface {
	// Get retrieves a metadata record by its ID.
	// The context parameter is used to control the lifetime of the request.
	// The id parameter is the unique identifier of the metadata record to retrieve.
	// It returns the metadata record and an error if the record is not found or there is another error.
	Get(ctx context.Context, id string) (*model.Metadata, error)
	Put(ctx context.Context, id string, m *model.Metadata) error
}

// Controller is a struct that holds a metadataRepository, which is used to interact with the metadata repository.
type Controller struct {
	repo metadataRepository
}

// New creates a new instance of the Controller struct, which holds a metadataRepository
// that is used to interact with the metadata repository.
func New(repo metadataRepository) *Controller {
	return &Controller{
		repo: repo,
	}
}

// Get retrieves a metadata record by its ID. The context parameter is used to control the lifetime of the request.
// The id parameter is the unique identifier of the metadata record to retrieve.
// if the record is not found or there is another error.
func (c *Controller) Get(ctx context.Context, id string) (*model.Metadata, error) {
	ctx, span := otel.Tracer("").Start(ctx, "GetController")
	defer span.End()

	res, err := c.repo.Get(ctx, id)
	if err != nil && errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Controller) Put(ctx context.Context, m *model.Metadata) error {
	ctx, span := otel.Tracer("").Start(ctx, "PutController")
	defer span.End()

	return c.repo.Put(ctx, m.ID, m)
}
