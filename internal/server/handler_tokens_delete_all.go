package server

import (
	"net/http"

	"github.com/urlspace/api/internal/user"
)

type tokenDeleteAllResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func handleTokensDeleteAll(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := userIDFromContext(r.Context())

		if err := svc.TokenDeleteAll(r.Context(), userID); err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusOK, tokenDeleteAllResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}
