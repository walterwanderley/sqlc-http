-- +goose Up
CREATE TABLE IF NOT EXISTS bios (
    id   integer    PRIMARY KEY AUTOINCREMENT,
    text text   NOT NULL
);

INSERT INTO bios (text) VALUES ('An American author known for his works of fiction and non-fiction.');
INSERT INTO bios (text) VALUES ('A British author famous for his fantasy novels and short stories.');
INSERT INTO bios (text) VALUES ('A Canadian author recognized for his contributions to science fiction and fantasy literature.');
INSERT INTO bios (text) VALUES ('A Brazilian author celebrated for his magical realism and social commentary in his works.');

CREATE TABLE IF NOT EXISTS authors (
    id   integer    PRIMARY KEY AUTOINCREMENT,
    name text   NOT NULL,
    bio  text,
    birth_date date,
    FOREIGN KEY(bio) REFERENCES bios(id)
);

-- +goose Down
DROP TABLE IF EXISTS authors;
DROP TABLE IF EXISTS bios;