package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/hreftools/api/internal/models"
	"github.com/hreftools/api/internal/response"
	"github.com/hreftools/api/internal/user"
)

type UsersListResponse struct {
	Status string                     `json:"status"`
	Data   []models.ResponseUserAdmin `json:"data"`
}

func UsersList(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := svc.List(r.Context())
		if err != nil {
			response.HandleDbError(w, err)
			return
		}

		items := make([]models.ResponseUserAdmin, len(list))
		for i, item := range list {
			items[i] = models.NewResponseUserAdmin(item)
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
