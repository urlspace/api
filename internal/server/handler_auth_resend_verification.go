package server

import (
	"encoding/json/v2"
	"net/http"

	"github.com/urlspace/api/internal/user"
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
		if err := json.UnmarshalRead(r.Body, &body, json.RejectUnknownMembers(true)); err != nil {
			handleClientError(r.Context(), w, err, "invalid request body")
			return
		}

		err := svc.ResendVerification(r.Context(), body.Email)
		if err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusOK, authResendVerificationResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}
