-- name: GetRatings :many
SELECT user_id, value
FROM ratings
WHERE record_id = $1
  AND record_type = $2;

-- name: InsertRating :execresult
INSERT INTO ratings (record_id, record_type, user_id, value)
VALUES ($1, $2, $3, $4);
