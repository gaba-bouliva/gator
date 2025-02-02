# Gator

Gator is a command-line tool for managing and aggregating RSS feeds. It allows users to register, log in, follow feeds, and browse posts from their followed feeds. Gator also supports periodic aggregation of feeds to keep the content up-to-date.

## Prerequisites

To run this program, you will need to have the following installed:

- [Go](https://golang.org/dl/)
- [PostgreSQL](https://www.postgresql.org/download/)
- [Goose](https://github.com/pressly/goose) for database migrations
- [SQLC](https://github.com/kyleconroy/sqlc) for generating type-safe code from SQL queries

## Installation

To install the `gator` CLI, use the following command:

```sh
go install github.com/gaba-bouliva/gator@latest
```

## Configuration

### Database Setup

1. Create the database and user in PostgreSQL:

```sql
CREATE DATABASE yourdbname;
CREATE USER yourusername WITH PASSWORD 'yourpassword';
GRANT ALL PRIVILEGES ON DATABASE yourdbname TO yourusername;
```

2. Create a configuration file named `.gatorconfig.json` in your home directory with the following content:

```json
{
  "db_url": "postgres://yourusername:yourpassword@localhost:5432/yourdbname?sslmode=disable",
  "current_user_name": ""
}
```

- `db_url`: The connection string for your PostgreSQL database.
- `current_user_name`: The username of the currently logged-in user (this will be set automatically when you log in).

3. Run the database migrations manually:

```sh
goose -dir sql/schema postgres "postgres://yourusername:yourpassword@localhost:5432/yourdbname?sslmode=disable" up
```

## Running the Program

To run the program, use the following command:

```sh
gator [command] [argument(s)]
```

## Commands

Here are a few commands you can run with the `gator` CLI:

- `gator login <username>`: Logs in a user.
- `gator register <username>`: Registers a new user.
- `gator reset`: Resets the user database.
- `gator users`: Lists all users.
- `gator agg <duration (1s, 1m, 1h)>`: Aggregates feeds at the specified interval.
- `gator addfeed <name> <url>`: Adds a new feed.
- `gator feeds`: Lists all feeds.
- `gator follow <url>`: Follows a feed by URL.
- `gator following`: Lists all followed feeds.
- `gator unfollow <url>`: Unfollows a feed by URL.
- `gator browse [limit (number)]`: Browses posts with an optional limit.
