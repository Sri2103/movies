package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // Import the PostgreSQL driver
	dbGen "movieexample.com/gen/db"

	config "movieexample.com/rating/configs"
	"movieexample.com/rating/internal/controller/rating"
	"movieexample.com/rating/internal/repository"
	"movieexample.com/rating/pkg/model"
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

func ConnectSQL(config *config.Config) (rating.RatingRepository, error) {
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

func (r *repo) Get(ctx context.Context, recordID model.RecordID, recordType model.RecordType) ([]model.Rating, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT user_id, value FROM ratings WHERE record_id = ? AND record_type = ?", recordID, recordType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []model.Rating
	for rows.Next() {
		var userID string
		var value int32
		if err := rows.Scan(&userID, &value); err != nil {
			return nil, err
		}
		res = append(res, model.Rating{
			UserID: model.UserID(userID),
			Value:  model.RatingValue(value),
		})
	}
	if len(res) == 0 {
		return nil, repository.ErrNotFound
	}
	return res, nil
}

// Put adds a rating for a given record.
func (r *repo) Put(ctx context.Context, recordID model.RecordID, recordType model.RecordType, rating *model.Rating) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO ratings (record_id, record_type, user_id, value) VALUES (?, ?, ?, ?)",
		recordID, recordType, rating.UserID, rating.Value)
	return err
}
