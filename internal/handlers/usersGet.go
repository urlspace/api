package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jumplist/api/internal/db"
	"github.com/jumplist/api/internal/response"
	"github.com/jumplist/api/internal/store"
)

type UsersGetResponse struct {
	Status string  `json:"status"`
	Data   db.User `json:"data"`
}

func UsersGet(store *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		idUuid, err := uuid.Parse(id)
		if err != nil {
			response.HandleClientError(w, err, "invalid id parameter")
			return
		}

		u, err := store.Users.GetById(r.Context(), idUuid)
		if err != nil {
			response.HandleDbError(w, err)
			return
		}

		response := &UsersGetResponse{
			Status: "ok",
			Data:   u,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
