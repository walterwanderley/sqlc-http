# Fullstack application go + sqlite + htmx

The most productive and efficient stack of the world!

## Steps to generate the code

0. Install the required tools.

```sh
go install github.com/walterwanderley/sqlc-http@latest
```
```sh
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

1. Create a directory to store SQL scripts.

```sh
mkdir -p sql/migrations
```

2. Create migrations scripts using [goose](https://github.com/pressly/goose?tab=readme-ov-file#migrations) rules.

```sh
echo "-- +goose Up
CREATE TABLE IF NOT EXISTS authors (
    id         integer PRIMARY KEY AUTOINCREMENT,
    name       text    NOT NULL,
    bio        text,
    birth_date date
);

-- +goose Down
DROP TABLE IF EXISTS authors;
" > sql/migrations/001_authors.sql
```

3. Create SQL queries and use [sqlc](https://sqlc.dev/) and [sqlc-http](http://github.com/walterwanderley/sqlc-http) comments syntax.

```sh
echo "/* name: GetAuthor :one */
/* http: GET /authors/{id}*/
SELECT * FROM authors
WHERE id = ? LIMIT 1;

/* name: ListAuthors :many */
/* http: GET /authors */
SELECT * FROM authors
ORDER BY name
LIMIT ? OFFSET ?;

/* name: CreateAuthor :execresult */
/* http: POST /authors */
INSERT INTO authors (
  name, bio, birth_date
) VALUES (
  ?, ?, ? 
);

/* name: UpdateAuthor :execresult */
/* http: PUT /authors/{id} */
UPDATE authors
SET name = ?, 
bio = ?,
birth_date = ?
WHERE id = ?;

/* name: UpdateAuthorBio :execresult */
/* http: PATCH /authors/{id}/bio */
UPDATE authors
SET bio = ?
WHERE id = ?;

/* name: DeleteAuthor :exec */
/* http: DELETE /authors/{id} */
DELETE FROM authors
WHERE id = ?;
" > sql/queries.sql
```

4. Create the sqlc.yaml configuration file

```sh
echo "
version: "2"
sql:
- schema: "./sql/migrations"
  queries: "./sql/queries.sql"
  engine: "sqlite"
  gen:
    go:
      out: "internal/authors"
" > sqlc.yaml
```

5. Execute sqlc

```sh
sqlc generate
```

6. Execute sqlc-http

```sh
sqlc-http -m sqlite-htmx -migration-path sql/migrations -frontend
```

## Running 

```sh
go run . -db test.db
```

Go to [http://localhost:5000](http://localhost:5000)

## Hot reload

If you want to automatic refresh the browser after change html files, use the **-dev** parameter:

```sh
go run . -db test.db -dev
```

"Computers make art, artists make money" (Chico Science)