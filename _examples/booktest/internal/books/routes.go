// Code generated by sqlc-http (https://github.com/walterwanderley/sqlc-http).

package books

import "net/http"

func (s *Service) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("POST /books/books-by-tags", s.handleBooksByTags())
	mux.HandleFunc("GET /books/books-by-title-year", s.handleBooksByTitleYear())
	mux.HandleFunc("POST /books/author", s.handleCreateAuthor())
	mux.HandleFunc("POST /books/book", s.handleCreateBook())
	mux.HandleFunc("DELETE /books/book/{book_id}", s.handleDeleteBook())
	mux.HandleFunc("GET /books/author/{author_id}", s.handleGetAuthor())
	mux.HandleFunc("GET /books/book/{book_id}", s.handleGetBook())
	mux.HandleFunc("PUT /books/book", s.handleUpdateBook())
	mux.HandleFunc("PUT /books/book-isbn", s.handleUpdateBookISBN())
}