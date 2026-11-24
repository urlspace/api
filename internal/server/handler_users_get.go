package server

import (
	"net/http"

	"github.com/urlspace/api/internal/user"
	"uuid"
)

type usersGetResponse struct {
	Status string            `json:"status"`
	Data   responseUserAdmin `json:"data"`
}

func handleUsersGet(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		idUuid, err := uuid.Parse(id)
		if err != nil {
			handleClientError(r.Context(), w, err, "invalid id parameter")
			return
		}

		u, err := svc.GetById(r.Context(), idUuid)
		if err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusOK, usersGetResponse{
			Status: "ok",
			Data:   newResponseUserAdmin(u),
		})
	}
}
