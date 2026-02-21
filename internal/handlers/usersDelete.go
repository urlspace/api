package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/hreftools/api/internal/db"
	"github.com/hreftools/api/internal/response"
	"github.com/hreftools/api/internal/store"
)

type UsersDeleteResponse struct {
	Status string  `json:"status"`
	Data   db.User `json:"data"`
}

func UsersDelete(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		idUuid, err := uuid.Parse(id)
		if err != nil {
			response.HandleClientError(w, err, "invalid id parameter")
			return
		}

		u, err := s.Users.GetById(r.Context(), idUuid)
		if err != nil {
			response.HandleDbError(w, err)
			return
		}

		err = s.Users.Delete(r.Context(), idUuid)
		if err != nil {
			response.HandleDbError(w, err)
			return
		}

		response := &UsersDeleteResponse{
			Status: "ok",
			Data:   u,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
