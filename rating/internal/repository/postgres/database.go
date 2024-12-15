package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // Import the PostgreSQL driver
	dbGen "movieexample.com/gen/db"

	config "movieexample.com/rating/configs"
	"movieexample.com/rating/internal/controller/rating"
	"movieexample.com/rating/pkg/model"
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

func CloseDB(r rating.Repository) {
	rp := r.(*repo)
	rp.db.Close()
}

func ConnectSQL(ctx context.Context, config *config.Config) (rating.Repository, error) {
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

func (r *repo) Put(ctx context.Context, recordID model.RecordID, recordType model.RecordType, rating *model.Rating) error {
	_, err := r.q.InsertRating(ctx, dbGen.InsertRatingParams{
		RecordID:   pgtype.Text{String: string(recordID), Valid: true},
		RecordType: pgtype.Text{String: string(recordType), Valid: true},
		UserID:     pgtype.Text{String: string(rating.UserID), Valid: true},
		Value:      pgtype.Int4{Int32: int32(rating.Value), Valid: true},
	})
	return err
}

func (r *repo) Get(ctx context.Context, recordID model.RecordID, recordType model.RecordType) ([]model.Rating, error) {
	data, err := r.q.GetRatings(ctx, dbGen.GetRatingsParams{
		RecordID:   pgtype.Text{String: string(recordID), Valid: true},
		RecordType: pgtype.Text{String: string(recordType), Valid: true},
	})
	if err != nil {
		return nil, err
	}
	var ratings []model.Rating
	for _, d := range data {
		ratings = append(ratings, model.Rating{
			UserID: model.UserID(d.UserID.String),
			Value:  model.RatingValue(d.Value.Int32),
		})
	}

	return ratings, nil
}
