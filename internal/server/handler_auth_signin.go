package server

import (
	"encoding/json"
	"net/http"

	"github.com/urlspace/api/internal/user"
)

type authSigninBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authSigninResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func handleAuthSignin(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body authSigninBody
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			handleClientError(r.Context(), w, err, "invalid request body")
			return
		}

		// No good reason for 1024 other than I like it.
		// Observe registered UA lengths before revisiting this limit.
		const maxUserAgentBytes = 1024
		ua := r.Header.Get("User-Agent")
		if len(ua) > maxUserAgentBytes {
			ua = ua[:maxUserAgentBytes]
		}
		var userAgent *string
		if ua != "" {
			userAgent = &ua
		}

		result, err := svc.Signin(r.Context(), body.Email, body.Password, userAgent)
		if err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		setSessionCookie(w, r, result.Session, result.ExpiresAt)

		writeJSONSuccess(w, http.StatusOK, authSigninResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}
