package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/hreftools/api/internal/config"
	"github.com/hreftools/api/internal/response"
	"github.com/hreftools/api/internal/user"
	"github.com/hreftools/api/internal/validator"
)

type AuthSigninBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (b *AuthSigninBody) Normalize() {
	b.Email = strings.ToLower(strings.TrimSpace(b.Email))
}

func (b *AuthSigninBody) Validate() error {
	if err := validator.Email(b.Email); err != nil {
		return err
	}

	if err := validator.Password(b.Password); err != nil {
		return err
	}

	return nil
}

type AuthSigninResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func AuthSignin(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body AuthSigninBody
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			response.HandleClientError(w, err, "invalid request body")
			return
		}

		body.Normalize()

		if err := body.Validate(); err != nil {
			response.HandleClientError(w, err, err.Error())
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
				response.WriteJSONError(w, http.StatusUnauthorized, "invalid email or password")
				return
			}
			response.HandleServerError(w, err, "failed to sign in")
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

		response.WriteJSONSuccess(w, http.StatusOK, AuthSigninResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}
