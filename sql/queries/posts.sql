-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at, published_at, title, description, url, feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING *;

-- name: GetPosts :many
SELECT * FROM posts ORDER BY published_at DESC LIMIT $1;

-- name: GetPostByUrl :one
SELECT * FROM posts WHERE url = $1;