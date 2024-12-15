-- name: GetMovie :one
SELECT title, description, director 
FROM movie 
WHERE id = $1;


-- name: InsertMovie :exec
INSERT INTO movie (id, title, description, director) 
VALUES ($1, $2, $3, $4);
