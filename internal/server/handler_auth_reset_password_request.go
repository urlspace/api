package server

import (
	"encoding/json/v2"
	"net/http"

	"github.com/urlspace/api/internal/user"
)

type authResetPasswordRequestBody struct {
	Email string `json:"email"`
}

type authResetPasswordRequestResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func handleAuthResetPasswordRequest(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body authResetPasswordRequestBody
		if err := json.UnmarshalRead(r.Body, &body, json.RejectUnknownMembers(true)); err != nil {
			handleClientError(r.Context(), w, err, "invalid request body")
			return
		}

		err := svc.ResetPasswordRequest(r.Context(), body.Email)
		if err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusOK, authResetPasswordRequestResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}
