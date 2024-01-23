// Code generated by sqlc-http (https://github.com/walterwanderley/sqlc-http). DO NOT EDIT.

package authors

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
)

type Service struct {
	querier *Queries
}

func NewService(querier *Queries) *Service {
	return &Service{querier: querier}
}

func (s *Service) handleCreateAuthor() http.HandlerFunc {
	type request struct {
		Name string  `json:"name"`
		Bio  *string `json:"bio"`
	}
	type response struct {
		LastInsertId int64 `json:"last_insert_id"`
		RowsAffected int64 `json:"rows_affected"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		var arg CreateAuthorParams
		arg.Name = req.Name
		if req.Bio != nil {
			arg.Bio = sql.NullString{Valid: true, String: *req.Bio}
		}

		result, err := s.querier.CreateAuthor(r.Context(), arg)
		if err != nil {
			slog.Error("sql call failed", "error", err, "method", "CreateAuthor")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		lastInsertId, _ := result.LastInsertId()
		rowsAffected, _ := result.RowsAffected()
		json.NewEncoder(w).Encode(response{
			LastInsertId: lastInsertId,
			RowsAffected: rowsAffected,
		})
	}
}

func (s *Service) handleDeleteAuthor() http.HandlerFunc {
	type request struct {
		Id int64 `json:"id"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		id := req.Id

		err := s.querier.DeleteAuthor(r.Context(), id)
		if err != nil {
			slog.Error("sql call failed", "error", err, "method", "DeleteAuthor")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s *Service) handleGetAuthor() http.HandlerFunc {
	type request struct {
		Id int64 `json:"id"`
	}
	type response struct {
		ID   int64   `json:"id"`
		Name string  `json:"name"`
		Bio  *string `json:"bio"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		id := req.Id

		result, err := s.querier.GetAuthor(r.Context(), id)
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
		json.NewEncoder(w).Encode(res)
	}
}

func (s *Service) handleListAuthors() http.HandlerFunc {
	type response struct {
		ID   int64   `json:"id"`
		Name string  `json:"name"`
		Bio  *string `json:"bio"`
	}

	return func(w http.ResponseWriter, r *http.Request) {

		result, err := s.querier.ListAuthors(r.Context())
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
		json.NewEncoder(w).Encode(res)
	}
}
