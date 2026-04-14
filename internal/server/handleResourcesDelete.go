package server

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/hreftools/api/internal/resource"
	"github.com/hreftools/api/internal/response"
	"github.com/hreftools/api/internal/utils"
)

type resourceDeleteResponse struct {
	Status string           `json:"status"`
	Data   responseResource `json:"data"`
}

func handleResourcesDelete(svc *resource.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := utils.UserIDFromContext(r.Context())

		id := r.PathValue("id")
		idUuid, err := uuid.Parse(id)
		if err != nil {
			response.HandleClientError(w, err, "invalid id parameter")
			return
		}

		rr, err := svc.Delete(r.Context(), idUuid, userID)
		if err != nil {
			response.HandleDbError(w, err)
			return
		}

		res := &resourceDeleteResponse{
			Status: "ok",
			Data:   newResponseResource(rr),
		}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
