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

-- name: GetFeedByURL :one
SELECT * FROM feeds
WHERE feeds.url = $1;

-- name: GetFeedUser :one
SELECT users.name FROM users
INNER JOIN feeds ON feeds.user_id = users.id
WHERE feeds.url = $1;

-- name: MarkFeedFetch :exec
UPDATE feeds
SET last_fetched_at = $1, updated_at = $1 
WHERE id = $2;

-- name: GetNextFeedToFetch :one
SELECT *
FROM feeds
ORDER BY feeds.last_fetched_at ASC NULLS FIRST
LIMIT 1;