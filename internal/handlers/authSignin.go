package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/hreftools/api/internal/response"
	"github.com/hreftools/api/internal/store"
	"github.com/hreftools/api/internal/utils"
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

func AuthSignin(s *store.Store) http.HandlerFunc {
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

		u, err := s.Users.GetByEmail(r.Context(), body.Email)
		if err != nil {
			response.WriteJSONError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}

		if !utils.PasswordValidate(body.Password, u.Password) {
			response.WriteJSONError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}

		if !u.EmailVerified {
			response.WriteJSONError(w, http.StatusForbidden, "email not verified")
			return
		}

		ua := r.Header.Get("User-Agent")
		var description *string
		if ua != "" {
			description = &ua
		}

		token, err := s.Tokens.Create(r.Context(), store.TokenCreateParams{
			UserID:      u.ID,
			Type:        TokenTypeSession,
			Description: description,
			ExpiresAt:   time.Now().Add(SessionExpiryDuration),
		})
		if err != nil {
			response.HandleServerError(w, err, "failed to create session")
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     SessionCookieName,
			Value:    token.ID.String(),
			Expires:  token.ExpiresAt,
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
