package server

import (
	"net/http"

	"github.com/urlspace/api/internal/user"
	"uuid"
)

type tokenDeleteResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func handleTokensDelete(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := userIDFromContext(r.Context())

		id := r.PathValue("id")
		idUuid, err := uuid.Parse(id)
		if err != nil {
			handleClientError(r.Context(), w, err, "invalid id parameter")
			return
		}

		if err := svc.TokenDelete(r.Context(), idUuid, userID); err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusOK, tokenDeleteResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}
