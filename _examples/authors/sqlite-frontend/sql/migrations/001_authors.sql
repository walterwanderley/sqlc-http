-- +goose Up
CREATE TABLE IF NOT EXISTS bios (
    id   integer    PRIMARY KEY AUTOINCREMENT,
    text text   NOT NULL
);

CREATE TABLE IF NOT EXISTS authors (
    id   integer    PRIMARY KEY AUTOINCREMENT,
    name text   NOT NULL,
    bio  text,
    birth_date date,
    FOREIGN KEY(bio) REFERENCES bios(id)
);

-- +goose Down
DROP TABLE IF EXISTS authors;