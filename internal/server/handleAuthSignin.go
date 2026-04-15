package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/hreftools/api/internal/config"
	"github.com/hreftools/api/internal/user"
	"github.com/hreftools/api/internal/validator"
)

type authSigninBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (b *authSigninBody) normalize() {
	b.Email = strings.ToLower(strings.TrimSpace(b.Email))
}

func (b *authSigninBody) validate() error {
	if err := validator.Email(b.Email); err != nil {
		return err
	}

	if err := validator.Password(b.Password); err != nil {
		return err
	}

	return nil
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

		body.normalize()

		if err := body.validate(); err != nil {
			handleClientError(w, err, err.Error())
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
			if errors.Is(err, user.ErrInvalidCredentials) || errors.Is(err, user.ErrEmailNotVerified) {
				writeJSONError(w, http.StatusUnauthorized, "invalid email or password")
				return
			}
			handleServerError(w, err, "failed to sign in")
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     config.SessionCookieName,
			Value:    result.Token.ID.String(),
			Expires:  result.Token.ExpiresAt,
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
