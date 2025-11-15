package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/zapi-sh/api/internal/db"
	"github.com/zapi-sh/api/internal/store"
)

type ResourcesListResponse struct {
	Status string        `json:"status"`
	Data   []db.Resource `json:"data"`
}

func ResourcesList(store *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := store.Resources.List(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		s := make([]db.Resource, 0, len(list))
		s = append(s, list...)

		response := &ResourcesListResponse{
			Status: "ok",
			Data:   s,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
