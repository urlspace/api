package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/zapi-sh/api/internal/db"
	"github.com/zapi-sh/api/internal/store"
)

type ResourceCreateBody struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Favourite   bool   `json:"favourite"`
	ReadLater   bool   `json:"read_later"`
}

type ResourceCreateResponse struct {
	Status string      `json:"status"`
	Data   db.Resource `json:"data"`
}

func ResourcesCreate(store *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body ResourceCreateBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			HandleClientError(w, err, "invalid request body")
			return
		}

		rr, err := store.Resources.Create(r.Context(), body.Title, body.Description, body.URL, body.Favourite, body.ReadLater)
		if err != nil {
			HandleDbError(w, err)
			return
		}

		response := &ResourceCreateResponse{
			Status: "ok",
			Data:   rr,
		}

		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
