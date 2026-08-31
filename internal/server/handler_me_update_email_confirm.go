package server

import (
	"encoding/json"
	"net/http"

	"github.com/urlspace/api/internal/user"
)

type meUpdateEmailConfirmBody struct {
	Code string `json:"code"`
}

type meUpdateEmailConfirmResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func handleMeUpdateEmailConfirm(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := userIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var body meUpdateEmailConfirmBody
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			handleClientError(r.Context(), w, err, "invalid request body")
			return
		}

		err := svc.ConfirmEmailChange(r.Context(), userID, body.Code)
		if err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusOK, meUpdateEmailConfirmResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}
