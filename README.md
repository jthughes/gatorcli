# Gator
## Description
Gator is a CLI app built in Go that allows multiple users to register, follow, and consume RSS feeds.

## Requirements
- Go: v1.23.4
- PostgreSQL: v15 or greater

## Installation
- ``go install ``
- Create ``~/.gatorconfig.json`` with the following fields:
```json
{
  "current_user_name": "",
  "db_url": "postgres://<username:password@url:port>/gator?sslmode=disable"
}
```

## Usage
- ``login <username>``: Login as ``<username>``.
- ``register <username>``: Register ``<username>`` as new username.
- ``agg <time_between_requests>``: Retrieves posts from all feeds on the specified duration, for example "10m30s".
- ``addfeed <feed_name> <feed_url>``: Add a new RSS feed url to the the lists of feeds that can be followed, and follows it for the current user.
- ``feeds``: Lists all feeds.
- ``follow <feed_url>``: Follows a feed that has already been registered by the ``addfeed`` command.
- ``following``: Lists all feeds that the current user is following.
- ``unfollow <feed_url>``: Unfollows a feed.