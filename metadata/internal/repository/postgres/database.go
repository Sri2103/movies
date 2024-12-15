package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // Import the PostgreSQL driver
	dbGen "movieexample.com/gen/db"
	config "movieexample.com/metadata/configs"
	"movieexample.com/metadata/internal/controller/metadata"
	"movieexample.com/metadata/internal/repository"
	"movieexample.com/metadata/pkg/model"
)

type repo struct {
	db *sql.DB
	q  dbGen.Queries
}

func NewDatabase(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func ConnectSQL(config *config.Config) (metadata.Repository, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		config.Postgres.Host,
		config.Postgres.Port,
		config.Postgres.Database,
		config.Postgres.Username,
		config.Postgres.Password,
		config.Postgres.SslMode,
	)
	db, err := NewDatabase(dsn)
	if err != nil {
		fmt.Println(err.Error(), "Error for creating database")
		panic(err)
	}
	db.SetMaxOpenConns(10)
	db.SetConnMaxIdleTime(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	return &repo{db: db, q: *dbGen.New(db)}, nil
}

// Get retrieves the metadata for the movie with the given ID from the database.
// If the movie is not found, it returns repository.ErrNotFound.
// If there is an error retrieving the metadata, it returns the error.
func (r *repo) Get(ctx context.Context, id string) (*model.Metadata, error) {
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
func (r *repo) Put(ctx context.Context, id string, metadata *model.Metadata) error {
	_, err := r.q.InsertMovie(ctx, dbGen.InsertMovieParams{
		ID:          id,
		Title:       sql.NullString{String: metadata.Title, Valid: metadata.Title != ""},
		Description: sql.NullString{String: metadata.Description, Valid: metadata.Description != ""},
		Director:    sql.NullString{String: metadata.Director, Valid: metadata.Director != ""},
	})
	return err
}
