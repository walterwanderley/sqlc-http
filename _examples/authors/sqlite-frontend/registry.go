// Code generated by sqlc-http (https://github.com/walterwanderley/sqlc-http). DO NOT EDIT.

package main

import (
	"database/sql"
	"net/http"

	authors_app "sqlite-htmx/internal/authors"
)

func registerHandlers(mux *http.ServeMux, db *sql.DB) {
	authorsService := authors_app.NewService(authors_app.New(db))
	authorsService.RegisterHandlers(mux)
}
