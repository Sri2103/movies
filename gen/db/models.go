// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package dbGen

import (
	"database/sql"
)

type Movie struct {
	ID          string
	Title       sql.NullString
	Description sql.NullString
	Director    sql.NullString
}

type Rating struct {
	RecordID   sql.NullString
	RecordType sql.NullString
	UserID     sql.NullString
	Value      sql.NullInt32
}
