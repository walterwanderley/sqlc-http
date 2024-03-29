// Code generated by sqlc-http (https://github.com/walterwanderley/sqlc-http). DO NOT EDIT.

package main

import (
	"database/sql"
	"net/http"

	books_app "booktest/internal/books"
)

func registerHandlers(mux *http.ServeMux, db *sql.DB) {
	booksService := books_app.NewService(books_app.New(db))
	booksService.RegisterHandlers(mux)
}
