package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql" // Import the MySQL driver
	dbGen "movieexample.com/gen/db"
	config "movieexample.com/metadata/configs"
	"movieexample.com/metadata/internal/controller/metadata"
	"movieexample.com/metadata/internal/repository"
	"movieexample.com/metadata/pkg/model"
)

type Repository struct {
	db *sql.DB
	q  dbGen.Queries
}

// New creates a new MySQL repository.
// It opens a connection to the MySQL database using the provided connection string.
// If the connection cannot be established, an error is returned.
func New(cfg *config.Config) (metadata.Repository, error) {
	dsn := fmt.Sprintf("%s:%s@/%s?allowPublicKeyRetrieval=%t&tls=%t&charset=utf8mb4&parseTime=true",
		"root", "rootPassword","test", true, false)

	fmt.Println(dsn, "DSN string")
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	// Ping the database to ensure the connection is established
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	log.Println("Connected to MySQL database", db.Stats())
	return &Repository{db: db, q: *dbGen.New(db)}, nil
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
	_, err := r.q.InsertMovie(ctx, dbGen.InsertMovieParams{
		ID:          id,
		Title:       sql.NullString{String: metadata.Title, Valid: metadata.Title != ""},
		Description: sql.NullString{String: metadata.Description, Valid: metadata.Description != ""},
		Director:    sql.NullString{String: metadata.Director, Valid: metadata.Director != ""},
	})
	return err
}

// func (r *Repository) Put(ctx context.Context, id string, metadata *model.Metadata) error {
// 	_, err := r.db.ExecContext(ctx, "INSERT INTO movies (id, title, description, director) VALUES (?, ?, ?, ?)",
// 		id, metadata.Title, metadata.Description, metadata.Director)
// 	return err
// }
