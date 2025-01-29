-- +goose Up
CREATE TABLE feeds (
    id SERIAL NOT NULL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    name VARCHAR(255) NOT NULL,
    url VARCHAR(255) NOT NULL UNIQUE,
    user_id INT NOT NULL,
    CONSTRAINT fk_user 
    FOREIGN KEY(user_id) REFERENCES users(ID)
    ON DELETE CASCADE
);

-- +goose Down
DROP TABLE feeds;