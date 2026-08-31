package server

import (
	"encoding/json"
	"net/http"

	"github.com/urlspace/api/internal/user"
)

type meUpdatePasswordBody struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

type meUpdatePasswordResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func handleMeUpdatePassword(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := userIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var body meUpdatePasswordBody
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			handleClientError(r.Context(), w, err, "invalid request body")
			return
		}

		err := svc.UpdatePassword(r.Context(), userID, body.CurrentPassword, body.NewPassword)
		if err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusOK, meUpdatePasswordResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}
