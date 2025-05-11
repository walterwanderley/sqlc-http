/* name: GetAuthor :one */
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
  name, bio
) VALUES (
  ?, ? 
);

/* name: UpdateAuthor :execresult */
/* http: PUT /authors/{id} */
UPDATE authors
SET name = ?, 
bio = ?
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