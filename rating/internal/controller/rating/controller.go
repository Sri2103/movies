package rating

import (
	"context"
	"errors"

	"movieexample.com/rating/internal/repository"
	"movieexample.com/rating/pkg/model"
)

// ErrNotFound is returned when a rating is not found for the given record.
var (
	ErrNotFound = errors.New("rating not found for the given record")
)

// ratingRepository is an interface that defines the methods for interacting with a rating storage system.
// The Get method retrieves a list of ratings for the given record ID and record type.
// The Put method stores a new rating for the given record ID and record type.
type ratingRepository interface {
	Get(ctx context.Context, recordID model.RecordID, recordType model.RecordType) ([]model.Rating, error)
	Put(ctx context.Context, recordID model.RecordID, recordType model.RecordType, rating *model.Rating) error
}

type rateIngester interface {
	Ingest(ctx context.Context) (chan model.RatingEvent, error)
}

// Controller is a struct that holds a ratingRepository, which is used to interact with a rating storage system.
type Controller struct {
	repo     ratingRepository
	ingester rateIngester
}

// NewController creates a new instance of the Controller struct with the provided ratingRepository.
func NewController(repo ratingRepository, ingester rateIngester) *Controller {
	return &Controller{
		repo:     repo,
		ingester: ingester,
	}
}

// GetAggregateRating retrieves the aggregate rating for the given record ID and record type. It calculates the average of all the ratings for the given record.
// If no ratings are found for the given record, it returns ErrNotFound.
func (c *Controller) GetAggregateRating(ctx context.Context, recordID model.RecordID, recordType model.RecordType) (float64, error) {
	ratings, err := c.repo.Get(ctx, recordID, recordType)
	if err != nil && errors.Is(err, repository.ErrNotFound) {
		return 0, ErrNotFound
	} else if err != nil {
		return 0, err
	}

	sum := float64(0)
	for _, r := range ratings {
		sum += float64(r.Value)
	}
	return sum / float64(len(ratings)), nil
}

// PutRating stores a new rating for the given record ID and record type.
func (c *Controller) PutRating(ctx context.Context, recordID model.RecordID, recordType model.RecordType, rating *model.Rating) error {
	return c.repo.Put(ctx, recordID, recordType, rating)
}

func (c *Controller) StartIngestion(ctx context.Context) error {
	ch, err := c.ingester.Ingest(ctx)
	if err != nil {
		return err
	}
	for e := range ch {
		if err := c.PutRating(ctx, e.RecordID, e.RecordType, &model.Rating{UserID: e.UserID, Value: e.Value}); err != nil {
			return err
		}
	}
	return nil
}
