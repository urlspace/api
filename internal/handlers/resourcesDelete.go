package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/zapi-sh/api/internal/db"
	"github.com/zapi-sh/api/internal/store"
)

type ResourceDeleteResponse struct {
	Status string      `json:"status"`
	Data   db.Resource `json:"data"`
}

func ResourcesDelete(store *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		idInt64, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			http.Error(w, "invalid id parameter", http.StatusBadRequest)
			return
		}

		rr, err := store.Resources.Get(r.Context(), idInt64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = store.Resources.Delete(r.Context(), idInt64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := &ResourcesGetResponse{
			Status: "ok",
			Data:   rr,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
