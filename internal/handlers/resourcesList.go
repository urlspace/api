package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/hreftools/api/internal/models"
	"github.com/hreftools/api/internal/response"
	"github.com/hreftools/api/internal/store"
	"github.com/hreftools/api/internal/utils"
)

type ResourcesListResponse struct {
	Status string                    `json:"status"`
	Data   []models.ResponseResource `json:"data"`
}

func ResourcesList(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := utils.UserIDFromContext(r.Context())

		list, err := s.Resources.List(r.Context(), userID)
		if err != nil {
			response.HandleDbError(w, err)
			return
		}

		items := make([]models.ResponseResource, len(list))
		for i, item := range list {
			items[i] = models.NewResponseResource(item)
		}

		response := &ResourcesListResponse{
			Status: "ok",
			Data:   items,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
