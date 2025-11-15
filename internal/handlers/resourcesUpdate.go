package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/zapi-sh/api/internal/db"
	"github.com/zapi-sh/api/internal/store"
)

type ResourceUpdateBody struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type ResourceUpdateResponse struct {
	Status string      `json:"status"`
	Data   db.Resource `json:"data"`
}

func ResourcesUpdate(store *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		id := r.PathValue("id")
		idInt64, err := strconv.ParseInt(id, 10, 64)
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

		rr, err := store.Resources.Update(r.Context(), idInt64, body.Title, body.URL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := &ResourceCreateResponse{
			Status: "ok",
			Data:   rr,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
