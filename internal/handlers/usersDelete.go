package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/hreftools/api/internal/models"
	"github.com/hreftools/api/internal/response"
	"github.com/hreftools/api/internal/user"
)

type UsersDeleteResponse struct {
	Status string                   `json:"status"`
	Data   models.ResponseUserAdmin `json:"data"`
}

func UsersDelete(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		idUuid, err := uuid.Parse(id)
		if err != nil {
			response.HandleClientError(w, err, "invalid id parameter")
			return
		}

		u, err := svc.Delete(r.Context(), idUuid)
		if err != nil {
			response.HandleDbError(w, err)
			return
		}

		response := &UsersDeleteResponse{
			Status: "ok",
			Data:   models.NewResponseUserAdmin(u),
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
