package server

import (
	"encoding/json"
	"net/http"

	"github.com/urlspace/api/internal/user"
)

type authSignupBody struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authSignupResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func handleAuthSignup(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body authSignupBody
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			handleClientError(w, err, "invalid request body")
			return
		}

		err := svc.Signup(r.Context(), body.Username, body.Email, body.Password)
		if err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusCreated, authSignupResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}
