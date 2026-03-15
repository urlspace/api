package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/hreftools/api/internal/models"
	"github.com/hreftools/api/internal/response"
	"github.com/hreftools/api/internal/store"
)

type UsersListResponse struct {
	Status string                `json:"status"`
	Data   []models.ResponseUser `json:"data"`
}

func UsersList(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := s.Users.List(r.Context())
		if err != nil {
			response.HandleDbError(w, err)
			return
		}

		items := make([]models.ResponseUser, len(list))
		for i, item := range list {
			items[i] = models.NewResponseUser(item)
		}

		response := &UsersListResponse{
			Status: "ok",
			Data:   items,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
