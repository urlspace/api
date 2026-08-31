package server

import (
	"encoding/json"
	"net/http"

	"github.com/urlspace/api/internal/user"
)

type meDeleteBody struct {
	Password string `json:"password"`
}

type meDeleteResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func handleMeDelete(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := userIDFromContext(r.Context())

		var body meDeleteBody
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			handleClientError(r.Context(), w, err, "invalid request body")
			return
		}

		if err := svc.DeleteSelf(r.Context(), userID, body.Password); err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		clearSessionCookie(w, r)

		writeJSONSuccess(w, http.StatusOK, meDeleteResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}
