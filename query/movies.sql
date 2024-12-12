-- name: GetMovie :one
SELECT
    *
FROM
    movie
WHERE
    "id" = ?;

-- name: InsertMovie :execresult
insert into
    movie (title, description, director)
values
    (?, ?, ?);