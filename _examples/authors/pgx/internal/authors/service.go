// Code generated by sqlc-http (https://github.com/walterwanderley/sqlc-http). DO NOT EDIT.

package authors

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"authors/internal/server"
)

type Service struct {
	querier Querier
	db      *pgxpool.Pool
}

func NewService(querier Querier, db *pgxpool.Pool) *Service {
	return &Service{querier: querier, db: db}
}

func (s *Service) handleCreateAuthor() http.HandlerFunc {
	type request struct {
		Name string  `form:"name" json:"name"`
		Bio  *string `form:"bio" json:"bio"`
	}
	type response struct {
		ID   int64   `json:"id,omitempty"`
		Name string  `json:"name,omitempty"`
		Bio  *string `json:"bio,omitempty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req, err := server.Decode[request](r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		arg := new(CreateAuthorParams)
		arg.Name = req.Name
		if req.Bio != nil {
			arg.Bio = pgtype.Text{Valid: true, String: *req.Bio}
		}

		result, err := s.querier.CreateAuthor(r.Context(), s.db, arg)
		if err != nil {
			slog.Error("sql call failed", "error", err, "method", "CreateAuthor")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var res response
		res.ID = result.ID
		res.Name = result.Name
		if result.Bio.Valid {
			res.Bio = &result.Bio.String
		}
		server.Encode(w, r, http.StatusOK, res)
	}
}

func (s *Service) handleDeleteAuthor() http.HandlerFunc {
	type request struct {
		Id int64 `form:"id" json:"id"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		if str := r.PathValue("id"); str != "" {
			if v, err := strconv.ParseInt(str, 10, 64); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			} else {
				req.Id = v
			}
		}
		id := req.Id

		err := s.querier.DeleteAuthor(r.Context(), s.db, id)
		if err != nil {
			slog.Error("sql call failed", "error", err, "method", "DeleteAuthor")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s *Service) handleGetAuthor() http.HandlerFunc {
	type request struct {
		Id int64 `form:"id" json:"id"`
	}
	type response struct {
		ID   int64   `json:"id,omitempty"`
		Name string  `json:"name,omitempty"`
		Bio  *string `json:"bio,omitempty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		if str := r.PathValue("id"); str != "" {
			if v, err := strconv.ParseInt(str, 10, 64); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			} else {
				req.Id = v
			}
		}
		id := req.Id

		result, err := s.querier.GetAuthor(r.Context(), s.db, id)
		if err != nil {
			slog.Error("sql call failed", "error", err, "method", "GetAuthor")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var res response
		res.ID = result.ID
		res.Name = result.Name
		if result.Bio.Valid {
			res.Bio = &result.Bio.String
		}
		server.Encode(w, r, http.StatusOK, res)
	}
}

func (s *Service) handleListAuthors() http.HandlerFunc {
	type response struct {
		ID   int64   `json:"id,omitempty"`
		Name string  `json:"name,omitempty"`
		Bio  *string `json:"bio,omitempty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {

		result, err := s.querier.ListAuthors(r.Context(), s.db)
		if err != nil {
			slog.Error("sql call failed", "error", err, "method", "ListAuthors")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		res := make([]response, 0)
		for _, r := range result {
			var item response
			item.ID = r.ID
			item.Name = r.Name
			if r.Bio.Valid {
				item.Bio = &r.Bio.String
			}
			res = append(res, item)
		}
		server.Encode(w, r, http.StatusOK, res)
	}
}
