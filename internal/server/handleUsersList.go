package server

import (
	"encoding/json"
	"net/http"

	"github.com/hreftools/api/internal/response"
	"github.com/hreftools/api/internal/user"
)

type usersListResponse struct {
	Status string              `json:"status"`
	Data   []responseUserAdmin `json:"data"`
}

func handleUsersList(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := svc.List(r.Context())
		if err != nil {
			response.HandleDbError(w, err)
			return
		}

		items := make([]responseUserAdmin, len(list))
		for i, item := range list {
			items[i] = newResponseUserAdmin(item)
		}

		res := &usersListResponse{
			Status: "ok",
			Data:   items,
		}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
