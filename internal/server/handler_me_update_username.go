package server

import (
	"encoding/json"
	"net/http"

	"github.com/urlspace/api/internal/user"
)

type meUpdateUsernameBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type meUpdateUsernameResponse struct {
	Status string       `json:"status"`
	Data   responseUser `json:"data"`
}

func handleMeUpdateUsername(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := userIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var body meUpdateUsernameBody
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			handleClientError(r.Context(), w, err, "invalid request body")
			return
		}

		u, err := svc.UpdateUsername(r.Context(), userID, body.Username, body.Password)
		if err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusOK, meUpdateUsernameResponse{
			Status: "ok",
			Data:   newResponseUser(u),
		})
	}
}
