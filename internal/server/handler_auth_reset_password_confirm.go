package server

import (
	"encoding/json/v2"
	"net/http"

	"github.com/urlspace/api/internal/user"
)

type authResetPasswordConfirmBody struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

type authResetPasswordConfirmResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func handleAuthResetPasswordConfirm(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body authResetPasswordConfirmBody
		if err := json.UnmarshalRead(r.Body, &body, json.RejectUnknownMembers(true)); err != nil {
			handleClientError(r.Context(), w, err, "invalid request body")
			return
		}

		err := svc.ResetPasswordConfirm(r.Context(), body.Token, body.Password)
		if err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusOK, authResetPasswordConfirmResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}
