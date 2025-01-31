-- +goose Up
CREATE TABLE feeds_follows (
    id SERIAL NOT NULL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    user_id INT NOT NULL,
    feeds_id INT NOT NULL,

    UNIQUE(user_id, feeds_id),

    CONSTRAINT fk_user
    FOREIGN KEY(user_id) REFERENCES users(id),
    CONSTRAINT fk_feeds
    FOREIGN KEY(feeds_id) REFERENCES feeds(id)
);

-- +goose Down
DROP TABLE feeds_follows;