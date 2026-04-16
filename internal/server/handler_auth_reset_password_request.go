package server

import (
	"encoding/json"
	"net/http"

	"github.com/hreftools/api/internal/user"
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
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			handleClientError(w, err, "invalid request body")
			return
		}

		err := svc.ResetPasswordRequest(r.Context(), body.Email)
		if err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusOK, authResetPasswordRequestResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}
