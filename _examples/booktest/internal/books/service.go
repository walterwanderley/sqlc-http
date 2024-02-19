// Code generated by sqlc-http (https://github.com/walterwanderley/sqlc-http). DO NOT EDIT.

package books

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"booktest/internal/server"
)

type Service struct {
	querier *Queries
}

func NewService(querier *Queries) *Service {
	return &Service{querier: querier}
}

func (s *Service) handleBooksByTags() http.HandlerFunc {
	type request struct {
		Dollar_1 []string `form:"dollar_1" json:"dollar_1"`
	}
	type response struct {
		BookID int32    `json:"book_id,omitempty"`
		Title  string   `json:"title,omitempty"`
		Name   *string  `json:"name,omitempty"`
		Isbn   string   `json:"isbn,omitempty"`
		Tags   []string `json:"tags,omitempty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req, err := server.Decode[request](r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		dollar_1 := req.Dollar_1

		result, err := s.querier.BooksByTags(r.Context(), dollar_1)
		if err != nil {
			slog.Error("sql call failed", "error", err, "method", "BooksByTags")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		res := make([]response, 0)
		for _, r := range result {
			var item response
			item.BookID = r.BookID
			item.Title = r.Title
			if r.Name.Valid {
				item.Name = &r.Name.String
			}
			item.Isbn = r.Isbn
			item.Tags = r.Tags
			res = append(res, item)
		}
		server.Encode(w, r, http.StatusOK, res)
	}
}

func (s *Service) handleBooksByTitleYear() http.HandlerFunc {
	type request struct {
		Title string `form:"title" json:"title"`
		Year  int32  `form:"year" json:"year"`
	}
	type response struct {
		BookID    int32     `json:"book_id,omitempty"`
		AuthorID  int32     `json:"author_id,omitempty"`
		Isbn      string    `json:"isbn,omitempty"`
		BookType  string    `json:"book_type,omitempty"`
		Title     string    `json:"title,omitempty"`
		Year      int32     `json:"year,omitempty"`
		Available time.Time `json:"available,omitempty"`
		Tags      []string  `json:"tags,omitempty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		req.Title = r.URL.Query().Get("title")
		if str := r.URL.Query().Get("year"); str != "" {
			if v, err := strconv.ParseInt(str, 10, 32); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			} else {
				req.Year = int32(v)
			}
		}
		var arg BooksByTitleYearParams
		arg.Title = req.Title
		arg.Year = req.Year

		result, err := s.querier.BooksByTitleYear(r.Context(), arg)
		if err != nil {
			slog.Error("sql call failed", "error", err, "method", "BooksByTitleYear")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		res := make([]response, 0)
		for _, r := range result {
			var item response
			item.BookID = r.BookID
			item.AuthorID = r.AuthorID
			item.Isbn = r.Isbn
			item.BookType = string(r.BookType)
			item.Title = r.Title
			item.Year = r.Year
			item.Available = r.Available
			item.Tags = r.Tags
			res = append(res, item)
		}
		server.Encode(w, r, http.StatusOK, res)
	}
}

func (s *Service) handleCreateAuthor() http.HandlerFunc {
	type request struct {
		Name string `form:"name" json:"name"`
	}
	type response struct {
		AuthorID int32  `json:"author_id,omitempty"`
		Name     string `json:"name,omitempty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req, err := server.Decode[request](r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		name := req.Name

		result, err := s.querier.CreateAuthor(r.Context(), name)
		if err != nil {
			slog.Error("sql call failed", "error", err, "method", "CreateAuthor")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var res response
		res.AuthorID = result.AuthorID
		res.Name = result.Name
		server.Encode(w, r, http.StatusOK, res)
	}
}

func (s *Service) handleCreateBook() http.HandlerFunc {
	type request struct {
		AuthorID  int32     `form:"author_id" json:"author_id"`
		Isbn      string    `form:"isbn" json:"isbn"`
		BookType  string    `form:"book_type" json:"book_type"`
		Title     string    `form:"title" json:"title"`
		Year      int32     `form:"year" json:"year"`
		Available time.Time `form:"available" json:"available"`
		Tags      []string  `form:"tags" json:"tags"`
	}
	type response struct {
		BookID    int32     `json:"book_id,omitempty"`
		AuthorID  int32     `json:"author_id,omitempty"`
		Isbn      string    `json:"isbn,omitempty"`
		BookType  string    `json:"book_type,omitempty"`
		Title     string    `json:"title,omitempty"`
		Year      int32     `json:"year,omitempty"`
		Available time.Time `json:"available,omitempty"`
		Tags      []string  `json:"tags,omitempty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req, err := server.Decode[request](r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		var arg CreateBookParams
		arg.AuthorID = req.AuthorID
		arg.Isbn = req.Isbn
		arg.BookType = BookType(req.BookType)
		arg.Title = req.Title
		arg.Year = req.Year
		arg.Available = req.Available
		arg.Tags = req.Tags

		result, err := s.querier.CreateBook(r.Context(), arg)
		if err != nil {
			slog.Error("sql call failed", "error", err, "method", "CreateBook")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var res response
		res.BookID = result.BookID
		res.AuthorID = result.AuthorID
		res.Isbn = result.Isbn
		res.BookType = string(result.BookType)
		res.Title = result.Title
		res.Year = result.Year
		res.Available = result.Available
		res.Tags = result.Tags
		server.Encode(w, r, http.StatusOK, res)
	}
}

func (s *Service) handleDeleteBook() http.HandlerFunc {
	type request struct {
		BookID int32 `form:"book_id" json:"book_id"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		if str := r.PathValue("book_id"); str != "" {
			if v, err := strconv.ParseInt(str, 10, 32); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			} else {
				req.BookID = int32(v)
			}
		}
		bookID := req.BookID

		err := s.querier.DeleteBook(r.Context(), bookID)
		if err != nil {
			slog.Error("sql call failed", "error", err, "method", "DeleteBook")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s *Service) handleGetAuthor() http.HandlerFunc {
	type request struct {
		AuthorID int32 `form:"author_id" json:"author_id"`
	}
	type response struct {
		AuthorID int32  `json:"author_id,omitempty"`
		Name     string `json:"name,omitempty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		if str := r.PathValue("author_id"); str != "" {
			if v, err := strconv.ParseInt(str, 10, 32); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			} else {
				req.AuthorID = int32(v)
			}
		}
		authorID := req.AuthorID

		result, err := s.querier.GetAuthor(r.Context(), authorID)
		if err != nil {
			slog.Error("sql call failed", "error", err, "method", "GetAuthor")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var res response
		res.AuthorID = result.AuthorID
		res.Name = result.Name
		server.Encode(w, r, http.StatusOK, res)
	}
}

func (s *Service) handleGetBook() http.HandlerFunc {
	type request struct {
		BookID int32 `form:"book_id" json:"book_id"`
	}
	type response struct {
		BookID    int32     `json:"book_id,omitempty"`
		AuthorID  int32     `json:"author_id,omitempty"`
		Isbn      string    `json:"isbn,omitempty"`
		BookType  string    `json:"book_type,omitempty"`
		Title     string    `json:"title,omitempty"`
		Year      int32     `json:"year,omitempty"`
		Available time.Time `json:"available,omitempty"`
		Tags      []string  `json:"tags,omitempty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		if str := r.PathValue("book_id"); str != "" {
			if v, err := strconv.ParseInt(str, 10, 32); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			} else {
				req.BookID = int32(v)
			}
		}
		bookID := req.BookID

		result, err := s.querier.GetBook(r.Context(), bookID)
		if err != nil {
			slog.Error("sql call failed", "error", err, "method", "GetBook")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var res response
		res.BookID = result.BookID
		res.AuthorID = result.AuthorID
		res.Isbn = result.Isbn
		res.BookType = string(result.BookType)
		res.Title = result.Title
		res.Year = result.Year
		res.Available = result.Available
		res.Tags = result.Tags
		server.Encode(w, r, http.StatusOK, res)
	}
}

func (s *Service) handleUpdateBook() http.HandlerFunc {
	type request struct {
		Title    string   `form:"title" json:"title"`
		Tags     []string `form:"tags" json:"tags"`
		BookType string   `form:"book_type" json:"book_type"`
		BookID   int32    `form:"book_id" json:"book_id"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req, err := server.Decode[request](r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		var arg UpdateBookParams
		arg.Title = req.Title
		arg.Tags = req.Tags
		arg.BookType = BookType(req.BookType)
		arg.BookID = req.BookID

		err = s.querier.UpdateBook(r.Context(), arg)
		if err != nil {
			slog.Error("sql call failed", "error", err, "method", "UpdateBook")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s *Service) handleUpdateBookISBN() http.HandlerFunc {
	type request struct {
		Title  string   `form:"title" json:"title"`
		Tags   []string `form:"tags" json:"tags"`
		BookID int32    `form:"book_id" json:"book_id"`
		Isbn   string   `form:"isbn" json:"isbn"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req, err := server.Decode[request](r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		var arg UpdateBookISBNParams
		arg.Title = req.Title
		arg.Tags = req.Tags
		arg.BookID = req.BookID
		arg.Isbn = req.Isbn

		err = s.querier.UpdateBookISBN(r.Context(), arg)
		if err != nil {
			slog.Error("sql call failed", "error", err, "method", "UpdateBookISBN")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
