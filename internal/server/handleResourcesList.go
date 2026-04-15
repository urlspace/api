package server

import (
	"encoding/json"
	"net/http"

	"github.com/hreftools/api/internal/resource"
	"github.com/hreftools/api/internal/utils"
)

type resourcesListResponse struct {
	Status string             `json:"status"`
	Data   []responseResource `json:"data"`
}

func handleResourcesList(svc *resource.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := utils.UserIDFromContext(r.Context())

		list, err := svc.List(r.Context(), userID)
		if err != nil {
			handleDbError(w, err)
			return
		}

		items := make([]responseResource, len(list))
		for i, item := range list {
			items[i] = newResponseResource(item)
		}

		res := &resourcesListResponse{
			Status: "ok",
			Data:   items,
		}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
