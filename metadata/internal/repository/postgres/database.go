package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // Import the PostgreSQL driver
	dbGen "movieexample.com/gen/db"
	config "movieexample.com/metadata/configs"
	"movieexample.com/metadata/internal/controller/metadata"
	"movieexample.com/metadata/pkg/model"
)

type repo struct {
	db *pgxpool.Pool
	q  dbGen.Queries
}

func NewDatabase(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	conn, err := pgxpool.New(ctx, dsn)
	// db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if err = conn.Ping(ctx); err != nil {
		return nil, err
	}

	return conn, nil
}

func ConnectSQL(ctx context.Context, config *config.Config) (metadata.Repository, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		config.Postgres.Host,
		config.Postgres.Port,
		config.Postgres.Database,
		config.Postgres.Username,
		config.Postgres.Password,
		config.Postgres.SslMode,
	)

	db, err := NewDatabase(ctx, dsn)
	if err != nil {
		fmt.Println(err.Error(), "Error for creating database")
		panic(err)
	}
	return &repo{db: db, q: *dbGen.New(db)}, nil
}

func CloseDB(r metadata.Repository) {
	rp := r.(*repo)
	rp.db.Close()
}

// Get retrieves the metadata for the movie with the given ID from the database.
// If the movie is not found, it returns repository.ErrNotFound.
// If there is an error retrieving the metadata, it returns the error.
func (r *repo) Get(ctx context.Context, id string) (*model.Metadata, error) {
	mv, err := r.q.GetMovie(ctx, id)
	if err != nil {
		return nil, err
	}

	return &model.Metadata{
		ID:          id,
		Title:       mv.Title.String,
		Description: mv.Description.String,
		Director:    mv.Director.String,
	}, nil
}

// Put adds movie metadata for a given movie id.
func (r *repo) Put(ctx context.Context, id string, metadata *model.Metadata) error {
	err := r.q.InsertMovie(ctx, dbGen.InsertMovieParams{
		ID:          id,
		Title:       pgtype.Text{String: metadata.Title, Valid: metadata.Title != ""},
		Description: pgtype.Text{String: metadata.Description, Valid: metadata.Description != ""},
		Director:    pgtype.Text{String: metadata.Director, Valid: metadata.Director != ""},
	})
	return err
}
