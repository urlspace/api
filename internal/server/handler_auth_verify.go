package server

import (
	"encoding/json/v2"
	"net/http"

	"github.com/urlspace/api/internal/user"
)

type authVerifyBody struct {
	Token string `json:"token"`
}

type authVerifyResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func handleAuthVerify(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body authVerifyBody
		if err := json.UnmarshalRead(r.Body, &body, json.RejectUnknownMembers(true)); err != nil {
			handleClientError(r.Context(), w, err, "invalid request body")
			return
		}

		err := svc.Verify(r.Context(), body.Token)
		if err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusOK, authVerifyResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}
