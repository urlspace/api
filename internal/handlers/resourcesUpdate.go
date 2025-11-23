package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/zapi-sh/api/internal/db"
	"github.com/zapi-sh/api/internal/store"
)

type ResourceUpdateBody struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Favourite   bool   `json:"favourite"`
	ReadLater   bool   `json:"read_later"`
}

type ResourceUpdateResponse struct {
	Status string      `json:"status"`
	Data   db.Resource `json:"data"`
}

func ResourcesUpdate(store *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		id := r.PathValue("id")
		idUuid, err := uuid.Parse(id)
		if err != nil {
			http.Error(w, "invalid id parameter", http.StatusBadRequest)
			return
		}

		var body ResourceUpdateBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			log.Printf("Error decoding request body: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		rr, err := store.Resources.Update(r.Context(), idUuid, body.Title, body.Description, body.URL, body.Favourite, body.ReadLater)
		if err != nil {
			HandleDbError(w, err)
			return
		}

		response := &ResourceUpdateResponse{
			Status: "ok",
			Data:   rr,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
