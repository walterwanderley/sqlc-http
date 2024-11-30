// Code generated by sqlc-http (https://github.com/walterwanderley/sqlc-http). DO NOT EDIT.

package authors

import (
	"database/sql"
	"log/slog"
	"net/http"
	"strconv"

	"authors/internal/server"
)

type Service struct {
	querier *Queries
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
		req, err := server.Decode[request](r)
		if err != nil {
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
		server.Encode(w, r, http.StatusOK, response{
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
		if str := r.PathValue("id"); str != "" {
			if v, err := strconv.ParseInt(str, 10, 64); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			} else {
				req.Id = v
			}
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
		server.Encode(w, r, http.StatusOK, res)
	}
}

func (s *Service) handleUpdateAuthor() http.HandlerFunc {
	type request struct {
		Name string  `json:"name"`
		Bio  *string `json:"bio"`
		ID   int64   `json:"id"`
	}
	type response struct {
		LastInsertId int64 `json:"last_insert_id"`
		RowsAffected int64 `json:"rows_affected"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req, err := server.Decode[request](r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		if str := r.PathValue("id"); str != "" {
			if v, err := strconv.ParseInt(str, 10, 64); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			} else {
				req.ID = v
			}
		}
		var arg UpdateAuthorParams
		arg.Name = req.Name
		if req.Bio != nil {
			arg.Bio = sql.NullString{Valid: true, String: *req.Bio}
		}
		arg.ID = req.ID

		result, err := s.querier.UpdateAuthor(r.Context(), arg)
		if err != nil {
			slog.Error("sql call failed", "error", err, "method", "UpdateAuthor")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		lastInsertId, _ := result.LastInsertId()
		rowsAffected, _ := result.RowsAffected()
		server.Encode(w, r, http.StatusOK, response{
			LastInsertId: lastInsertId,
			RowsAffected: rowsAffected,
		})
	}
}

func (s *Service) handleUpdateAuthorBio() http.HandlerFunc {
	type request struct {
		Bio *string `json:"bio"`
		ID  int64   `json:"id"`
	}
	type response struct {
		LastInsertId int64 `json:"last_insert_id"`
		RowsAffected int64 `json:"rows_affected"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req, err := server.Decode[request](r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		if str := r.PathValue("id"); str != "" {
			if v, err := strconv.ParseInt(str, 10, 64); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			} else {
				req.ID = v
			}
		}
		var arg UpdateAuthorBioParams
		if req.Bio != nil {
			arg.Bio = sql.NullString{Valid: true, String: *req.Bio}
		}
		arg.ID = req.ID

		result, err := s.querier.UpdateAuthorBio(r.Context(), arg)
		if err != nil {
			slog.Error("sql call failed", "error", err, "method", "UpdateAuthorBio")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		lastInsertId, _ := result.LastInsertId()
		rowsAffected, _ := result.RowsAffected()
		server.Encode(w, r, http.StatusOK, response{
			LastInsertId: lastInsertId,
			RowsAffected: rowsAffected,
		})
	}
}
