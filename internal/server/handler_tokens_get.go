package server

import (
	"net/http"

	"github.com/urlspace/api/internal/user"
	"uuid"
)

type tokensGetResponse struct {
	Status string        `json:"status"`
	Data   responseToken `json:"data"`
}

func handleTokensGet(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := userIDFromContext(r.Context())

		id := r.PathValue("id")
		idUuid, err := uuid.Parse(id)
		if err != nil {
			handleClientError(r.Context(), w, err, "invalid id parameter")
			return
		}

		t, err := svc.TokenGetByID(r.Context(), idUuid, userID)
		if err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusOK, tokensGetResponse{
			Status: "ok",
			Data:   newResponseToken(t),
		})
	}
}
