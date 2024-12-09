package memory

import (
	"context"

	"movieexample.com/rating/internal/repository"
	"movieexample.com/rating/pkg/model"
)

// Repository is an in-memory implementation of the rating.Repository interface.
// It stores ratings in a nested map, with the outer map keyed by RecordType
// and the inner map keyed by RecordID, storing a slice of Rating values.
type Repository struct {
	data map[model.RecordType]map[model.RecordID][]model.Rating
}

// New returns a new in-memory implementation of the rating.Repository interface.
// It initializes the data map to store ratings by RecordType and RecordID.
func New() *Repository {
	return &Repository{
		data: map[model.RecordType]map[model.RecordID][]model.Rating{},
	}
}

// Get retrieves the ratings for the specified record ID and record type. If no
// ratings are found for the given record, it returns an ErrNotFound error.
func (r *Repository) Get(_ context.Context, recordID model.RecordID, recordType model.RecordType) ([]model.Rating, error) {
	if _, ok := r.data[recordType]; !ok {
		return nil, repository.ErrNotFound
	}
	if ratings, ok := r.data[recordType][recordID]; !ok || len(ratings) == 0 {
		return nil, repository.ErrNotFound
	}
	return r.data[recordType][recordID], nil
}

// Put stores the provided rating for the specified record ID and record type.
// If the record type or record ID does not exist in the repository, it will
// create new entries for them.
func (r *Repository) Put(_ context.Context, recordID model.RecordID, recordType model.RecordType, rating *model.Rating) error {
	if _, ok := r.data[recordType]; !ok {
		r.data[recordType] = map[model.RecordID][]model.Rating{}
	}
	r.data[recordType][recordID] = append(r.data[recordType][recordID], *rating)
	return nil
}
