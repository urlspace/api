package server

import (
	"encoding/json"
	"net/http"

	"github.com/urlspace/api/internal/config"
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
			handleClientError(w, err, "invalid request body")
			return
		}

		const maxUaLength = 255
		ua := r.Header.Get("User-Agent")
		if len(ua) > maxUaLength {
			ua = ua[:maxUaLength]
		}
		var description *string
		if ua != "" {
			description = &ua
		}

		result, err := svc.Signin(r.Context(), body.Email, body.Password, description)
		if err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     config.SessionCookieName,
			Value:    result.Session.ID.String(),
			Expires:  result.Session.ExpiresAt,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		})

		writeJSONSuccess(w, http.StatusOK, authSigninResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}
