package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/hreftools/api/internal/db"
	"github.com/hreftools/api/internal/response"
	"github.com/hreftools/api/internal/store"
)

type ResourcesListResponse struct {
	Status string        `json:"status"`
	Data   []db.Resource `json:"data"`
}

func ResourcesList(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := s.Resources.List(r.Context())
		if err != nil {
			response.HandleDbError(w, err)
			return
		}

		response := &ResourcesListResponse{
			Status: "ok",
			Data:   list,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
