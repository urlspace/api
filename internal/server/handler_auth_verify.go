package server

import (
	"encoding/json"
	"net/http"

	"github.com/hreftools/api/internal/user"
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
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			handleClientError(w, err, "invalid request body")
			return
		}

		err := svc.Verify(r.Context(), body.Token)
		if err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusOK, authVerifyResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}
