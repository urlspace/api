package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/hreftools/api/internal/db"
	"github.com/hreftools/api/internal/response"
	"github.com/hreftools/api/internal/store"
)

type UsersListResponse struct {
	Status string    `json:"status"`
	Data   []db.User `json:"data"`
}

func UsersList(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := s.Users.List(r.Context())
		if err != nil {
			response.HandleDbError(w, err)
			return
		}

		response := &UsersListResponse{
			Status: "ok",
			Data:   list,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
