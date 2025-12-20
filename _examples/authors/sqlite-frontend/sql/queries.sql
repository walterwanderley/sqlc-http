/* name: GetAuthor :one */
/* http: GET /authors/{id}*/
/* ref: bio listBios */
SELECT * FROM authors
WHERE id = ? LIMIT 1;

/* name: ListAuthors :many */
/* http: GET /authors */
SELECT * FROM authors
ORDER BY name
LIMIT ? OFFSET ?;

-- name: listBios :many
SELECT id, text FROM bios;

/* name: CreateAuthor :execresult */
/* http: POST /authors */
/* ref: bio listBios */
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