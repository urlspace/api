package server

import (
	"net/http"

	"github.com/urlspace/api/internal/user"
)

type sessionDeleteAllResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func handleSessionsDeleteAll(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := userIDFromContext(r.Context())

		if err := svc.SessionDeleteAll(r.Context(), userID); err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusOK, sessionDeleteAllResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}
