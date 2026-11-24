package server

import (
	"net/http"

	"github.com/urlspace/api/internal/user"
)

type usersListResponse struct {
	Status string              `json:"status"`
	Data   []responseUserAdmin `json:"data"`
}

func handleUsersList(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := svc.List(r.Context())
		if err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		items := make([]responseUserAdmin, len(list))
		for i, item := range list {
			items[i] = newResponseUserAdmin(item)
		}

		writeJSONSuccess(w, http.StatusOK, usersListResponse{
			Status: "ok",
			Data:   items,
		})
	}
}
