package memory

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel"
	"movieexample.com/metadata/internal/repository"
	"movieexample.com/metadata/pkg/model"
)

type Repository struct {
	sync.RWMutex
	data map[string]*model.Metadata
}

// New returns a new in-memory repository for storing Metadata.
func New() *Repository {
	return &Repository{
		data: map[string]*model.Metadata{},
	}
}

// Get retrieves the Metadata for the given id from the in-memory repository.
// If the Metadata is not found, it returns repository.ErrNotFound.
func (r *Repository) Get(ctx context.Context, id string) (*model.Metadata, error) {
	_, span := otel.Tracer("").Start(ctx, "GetMemoryRepo")
	defer span.End()
	r.RLock()
	defer r.RUnlock()

	m, ok := r.data[id]

	if !ok {
		return nil, repository.ErrNotFound
	}

	return m, nil
}

// Put stores the given Metadata in the in-memory repository, keyed by the Metadata's ID.
// If the Metadata already exists, it will be overwritten.
func (r *Repository) Put(_ context.Context, id string, m *model.Metadata) error {
	r.Lock()
	defer r.Unlock()
	r.data[id] = m

	return nil
}
