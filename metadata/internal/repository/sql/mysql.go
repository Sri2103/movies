package mysql

import (
	"context"
	"database/sql"

	_ "github.com/go-sql-driver/mysql" // Import the MySQL driver
	"movieexample.com/metadata/internal/repository"
	"movieexample.com/metadata/pkg/model"
)

type Repository struct {
	db *sql.DB
}

// New creates a new MySQL repository.
// It opens a connection to the MySQL database using the provided connection string.
// If the connection cannot be established, an error is returned.
func New() (*Repository, error) {
	db, err := sql.Open("mysql", "root:password@/movieexample")
	if err != nil {
		return nil, err
	}
	return &Repository{db: db}, nil
}

// Get retrieves the metadata for the movie with the given ID from the database.
// If the movie is not found, it returns repository.ErrNotFound.
// If there is an error retrieving the metadata, it returns the error.
func (r *Repository) Get(ctx context.Context, id string) (*model.Metadata, error) {
	var title, description, director string
	err := r.db.QueryRowContext(ctx, "SELECT title, description, director FROM movie WHERE id = ?", id).Scan(&title, &description, &director)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &model.Metadata{
		ID:          id,
		Title:       title,
		Description: description,
		Director:    director,
	}, nil
}

// Put adds movie metadata for a given movie id.
func (r *Repository) Put(ctx context.Context, id string, metadata *model.Metadata) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO movies (id, title, description, director) VALUES (?, ?, ?, ?)",
		id, metadata.Title, metadata.Description, metadata.Director)
	return err
}
