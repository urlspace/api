package server

import (
	"net/http"

	"github.com/urlspace/api/internal/user"
	"uuid"
)

type sessionDeleteResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func handleSessionsDelete(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := userIDFromContext(r.Context())

		id := r.PathValue("id")
		idUuid, err := uuid.Parse(id)
		if err != nil {
			handleClientError(r.Context(), w, err, "invalid id parameter")
			return
		}

		if err := svc.SessionDelete(r.Context(), idUuid, userID); err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusOK, sessionDeleteResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}
