package server

import (
	"encoding/json"
	"net/http"

	"github.com/hreftools/api/internal/user"
)

type authResendVerificationBody struct {
	Email string `json:"email"`
}

type authResendVerificationResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func handleAuthResendVerification(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body authResendVerificationBody
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			handleClientError(w, err, "invalid request body")
			return
		}

		err := svc.ResendVerification(r.Context(), body.Email)
		if err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusOK, authResendVerificationResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}
