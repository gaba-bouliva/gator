-- +goose Up
CREATE TABLE posts (
    id SERIAL NOT NULL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    published_at TIMESTAMP NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    url VARCHAR(255) NOT NULL UNIQUE,
    feed_id INT NOT NULL,
    CONSTRAINT fk_feed
    FOREIGN KEY(feed_id) REFERENCES feeds(ID)
);

-- +goose Down
DROP TABLE posts;