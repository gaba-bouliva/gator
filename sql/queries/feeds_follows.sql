-- name: CreateFeedFollow :one
WITH inserted AS (
    INSERT INTO feeds_follows (id, created_at, updated_at, user_id, feeds_id)
    VALUES (
        $1,
        $2,
        $3,
        $4,
        $5
    )
    RETURNING id, created_at, updated_at, user_id, feeds_id
)
SELECT 
  inserted.id,
  inserted.created_at,
  inserted.updated_at,
  inserted.user_id,
  inserted.feeds_id,
  users.name AS user_name,
  feeds.name AS feed_name
FROM inserted
JOIN users ON inserted.user_id = users.id
JOIN feeds ON inserted.feeds_id = feeds.id;

-- name: GetFeedFollowsForUser :many
SELECT 
    feeds_follows.id, 
    feeds_follows.created_at, 
    feeds_follows.updated_at, 
    feeds_follows.user_id, 
    feeds_follows.feeds_id, 
    users.name AS user_name, 
    feeds.name AS feed_name
FROM feeds_follows
JOIN users ON users.id = feeds_follows.user_id
JOIN feeds ON feeds.id = feeds_follows.feeds_id
WHERE feeds_follows.user_id = $1;
 