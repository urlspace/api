package server

import (
	"encoding/json"
	"net/http"

	"github.com/urlspace/api/internal/user"
)

type tokenCreateBody struct {
	Password    string `json:"password"`
	Description string `json:"description"`
}

type tokenCreateResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func handleTokensCreate(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := userIDFromContext(r.Context())

		var body tokenCreateBody
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			handleClientError(w, err, "invalid request body")
			return
		}

		result, err := svc.TokenCreate(r.Context(), userID, body.Password, body.Description)
		if err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusCreated, tokenCreateResponse{
			Status: "ok",
			Data:   result.RawToken,
		})
	}
}
