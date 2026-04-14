package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/hreftools/api/internal/models"
	"github.com/hreftools/api/internal/response"
	"github.com/hreftools/api/internal/user"
)

type UsersGetResponse struct {
	Status string                   `json:"status"`
	Data   models.ResponseUserAdmin `json:"data"`
}

func UsersGet(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		idUuid, err := uuid.Parse(id)
		if err != nil {
			response.HandleClientError(w, err, "invalid id parameter")
			return
		}

		u, err := svc.GetById(r.Context(), idUuid)
		if err != nil {
			response.HandleDbError(w, err)
			return
		}

		response := &UsersGetResponse{
			Status: "ok",
			Data:   models.NewResponseUserAdmin(u),
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
