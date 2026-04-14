package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/hreftools/api/internal/models"
	"github.com/hreftools/api/internal/resource"
	"github.com/hreftools/api/internal/response"
	"github.com/hreftools/api/internal/utils"
)

type ResourcesGetResponse struct {
	Status string                  `json:"status"`
	Data   models.ResponseResource `json:"data"`
}

func ResourcesGet(svc *resource.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := utils.UserIDFromContext(r.Context())

		id := r.PathValue("id")
		idUuid, err := uuid.Parse(id)
		if err != nil {
			response.HandleClientError(w, err, "invalid id parameter")
			return
		}

		rr, err := svc.Get(r.Context(), idUuid, userID)
		if err != nil {
			response.HandleDbError(w, err)
			return
		}

		response := &ResourcesGetResponse{
			Status: "ok",
			Data:   models.NewResponseResource(rr),
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
