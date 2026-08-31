package server

import (
	"net/http"

	"github.com/urlspace/api/internal/user"
)

type sessionsListResponse struct {
	Status string             `json:"status"`
	Data   []responseSession `json:"data"`
}

func handleSessionsList(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := userIDFromContext(r.Context())

		list, err := svc.SessionList(r.Context(), userID)
		if err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		currentSessionID, _ := sessionIDFromContext(r.Context())

		items := make([]responseSession, len(list))
		for i, item := range list {
			items[i] = newResponseSession(item, currentSessionID)
		}

		writeJSONSuccess(w, http.StatusOK, sessionsListResponse{
			Status: "ok",
			Data:   items,
		})
	}
}
