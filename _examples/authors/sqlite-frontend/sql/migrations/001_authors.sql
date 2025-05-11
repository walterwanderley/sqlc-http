-- +goose Up
CREATE TABLE IF NOT EXISTS authors (
    id   integer    PRIMARY KEY AUTOINCREMENT,
    name text   NOT NULL,
    bio  text,
    created_at date
);

-- +goose Down
DROP TABLE IF EXISTS authors;