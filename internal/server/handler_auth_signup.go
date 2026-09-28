package server

import (
	"encoding/json/v2"
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
		if err := json.UnmarshalRead(r.Body, &body, json.RejectUnknownMembers(true)); err != nil {
			handleClientError(r.Context(), w, err, "invalid request body")
			return
		}

		err := svc.Signup(r.Context(), body.Username, body.Email, body.Password)
		if err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusCreated, authSignupResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}
