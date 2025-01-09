-- name: AddFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: GetFeeds :many
SELECT * FROM feeds;

-- name: GetFeedUser :one
SELECT users.name FROM users
INNER JOIN feeds ON feeds.user_id = users.id
WHERE feeds.url = $1;