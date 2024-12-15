-- name: GetRatings :many
SELECT
    user_id,
    value
FROM
    ratings
WHERE
    record_id = ?
    AND record_type = ?;

-- name: InsertRating :execresult
INSERT INTO
    ratings (record_id, record_type, user_id, value)
VALUES
    (?, ?, ?, ?)