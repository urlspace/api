package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jumplist/api/internal/db"
	"github.com/jumplist/api/internal/response"
	"github.com/jumplist/api/internal/store"
)

type ResourceDeleteResponse struct {
	Status string      `json:"status"`
	Data   db.Resource `json:"data"`
}

func ResourcesDelete(store *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		idUuid, err := uuid.Parse(id)
		if err != nil {
			response.HandleClientError(w, err, "invalid id parameter")
			return
		}

		rr, err := store.Resources.Get(r.Context(), idUuid)
		if err != nil {
			response.HandleDbError(w, err)
			return
		}

		err = store.Resources.Delete(r.Context(), idUuid)
		if err != nil {
			response.HandleDbError(w, err)
			return
		}

		response := &ResourceDeleteResponse{
			Status: "ok",
			Data:   rr,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
