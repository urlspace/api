package server

import (
	"encoding/json/v2"
	"net/http"

	"github.com/urlspace/api/internal/user"
)

type meUpdateDisplayNameBody struct {
	DisplayName string `json:"displayName"`
}

type meUpdateDisplayNameResponse struct {
	Status string       `json:"status"`
	Data   responseUser `json:"data"`
}

func handleMeUpdateDisplayName(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := getUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var body meUpdateDisplayNameBody
		if err := json.UnmarshalRead(r.Body, &body, json.RejectUnknownMembers(true)); err != nil {
			handleClientError(r.Context(), w, err, "invalid request body")
			return
		}

		u, err := svc.UpdateDisplayName(r.Context(), userID, body.DisplayName)
		if err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusOK, meUpdateDisplayNameResponse{
			Status: "ok",
			Data:   newResponseUser(u),
		})
	}
}
