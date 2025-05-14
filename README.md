## sqlc-http

Create a **net/http Server** from the generated code by the awesome [sqlc](https://sqlc.dev/) project. If you’re searching for a SQLC plugin, use [sqlc-gen-go-server](https://github.com/walterwanderley/sqlc-gen-go-server/).

### Requirements

- Go 1.23 or superior
- [sqlc](https://sqlc.dev/)

```sh
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

### Installation

```sh
go install github.com/walterwanderley/sqlc-http@latest
```

### Fullstack application example (HTMX)

If you want to generate a complete application (including htmx frontend), check [this example](https://github.com/walterwanderley/sqlc-http/blob/main/_examples/authors/sqlite-frontend/README.md).

### Example

1. Create a queries.sql file:

```sql
--queries.sql

CREATE TABLE authors (
  id   BIGSERIAL PRIMARY KEY,
  name text      NOT NULL,
  bio  text,
  created_at TIMESTAMP
);

-- name: GetAuthor :one
-- http: GET /authors/{id}
SELECT * FROM authors
WHERE id = $1 LIMIT 1;

-- name: ListAuthors :many
-- http: GET /authors
SELECT * FROM authors
ORDER BY name;

-- name: CreateAuthor :one
-- http: POST /authors
INSERT INTO authors (
  name, bio, created_at
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: DeleteAuthor :exec
-- http: DELETE /authors/{id}
DELETE FROM authors
WHERE id = $1;

-- name: UpdateAuthorBio :exec
-- http: PATCH /authors/{id}/bio
UPDATE authors
SET bio = $1
WHERE id = $2;
```

2. Create a sqlc.yaml file

```yaml
version: "2"
sql:
- schema: "./queries.sql"
  queries: "./queries.sql"
  engine: "postgresql"
  gen:
    go:
      out: "internal/author"
```

3. Execute sqlc

```sh
sqlc generate
```

4. Execute sqlc-http

```sh
sqlc-http -m "mymodule"
```

If you want to generate the frontend (htmx):

```sh
sqlc-http -m "mymodule" -frontend
```


5. Run the generated server

```sh
go run . -db [Database Connection URL] -dev
```

6. Enjoy!

If you do not generate the frontend in step 4?

- Swagger UI: [http://localhost:5000/swagger](http://localhost:5000/swagger)

If you generate the frontend in step 4:

- [HTMX](https://htmx.org) frontend: [http://localhost:5000](http://localhost:5000)
- Swagger UI: [http://localhost:5000/web/swagger](http://localhost:5000/web/swagger)

### Customizing HTTP endpoints

You can customize the HTTP endpoints by adding comments to the queries.

```sql
-- http: Method Path
```

Here’s an example of a queries file that has a custom HTTP endpoint:
```sql
-- name: ListAuthors :many
-- http: GET /authors
SELECT * FROM authors
ORDER BY name;

-- name: UpdateAuthorBio :exec
-- http: PATCH /authors/{id}/bio
UPDATE authors
SET bio = $1
WHERE id = $2;
```


### Editing the generated code

- It's safe to edit any generated code that doesn't have the `DO NOT EDIT` indication at the very first line.

- After modify a SQL file, execute these commands below:

```sh
sqlc generate
go generate
```

### Similar Projects

- [sqlc-connect](https://github.com/walterwanderley/sqlc-connect)
- [sqlc-grpc](https://github.com/walterwanderley/sqlc-grpc)
- [xo-grpc](https://github.com/walterwanderley/xo-grpc)