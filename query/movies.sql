-- name: GetMovie :one
SELECT
    *
FROM
    Movie
WHERE
    "id" = ?;

-- name: InsertMovie :execresult
insert into
    Movie (id, title, description, director)
values
    (?, ?, ?, ?);