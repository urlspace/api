package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/hreftools/api/internal/models"
	"github.com/hreftools/api/internal/response"
	"github.com/hreftools/api/internal/user"
	"github.com/hreftools/api/internal/utils"
)

type MeGetResponse struct {
	Status string              `json:"status"`
	Data   models.ResponseUser `json:"data"`
}

func MeGet(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := utils.UserIDFromContext(r.Context())
		if !ok {
			response.WriteJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		u, err := svc.GetById(r.Context(), userID)
		if err != nil {
			response.HandleDbError(w, err)
			return
		}

		response := &MeGetResponse{
			Status: "ok",
			Data:   models.NewResponseUser(u),
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
