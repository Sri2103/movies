package model

// RecordTypeMovie is a constant representing the "movie" record type.
const (
	RecordTypeMovie = RecordType("movie")
)

// RecordID is a string type used to represent a unique identifier for a record.
type RecordID string

// RecordType is a string type used to represent the type of a record.
type RecordType string

// UserID is a string type used to represent a unique identifier for a user.
type UserID string

// RatingValue is an integer type used to represent a rating value.
type RatingValue int

// Rating represents a user's rating for a movie.
type Rating struct {
	ID      RecordID    `json:"id"`
	UserID  UserID      `json:"user_id"`
	MovieID RecordID    `json:"movie_id"`
	Value   RatingValue `json:"rating"`
}

type RatingEventType string

const (
	RatingEventPut    RatingEventType = "put"
	RatingEventDelete RatingEventType = "delete"
)

type RatingEvent struct {
	UserID     UserID      `json:"userId"`
	RecordID   RecordID    `json:"recordId"`
	RecordType RecordType  `json:"recordType"`
	Value      RatingValue `json:"value"`
}
