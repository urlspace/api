package server

import (
	"encoding/json"
	"net/http"

	"github.com/urlspace/api/internal/user"
)

type meUpdateEmailBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type meUpdateEmailResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func handleMeUpdateEmail(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := userIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var body meUpdateEmailBody
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			handleClientError(r.Context(), w, err, "invalid request body")
			return
		}

		err := svc.RequestEmailChange(r.Context(), userID, body.Email, body.Password)
		if err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusOK, meUpdateEmailResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}
