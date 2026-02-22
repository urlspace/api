package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hreftools/api/internal/models"
	"github.com/hreftools/api/internal/response"
	"github.com/hreftools/api/internal/store"
)

type AuthVerifyBody struct {
	Token string `json:"token"`
}

func (b *AuthVerifyBody) Validate() error {
	b.Token = strings.TrimSpace(b.Token)

	if len(b.Token) == 0 {
		return errors.New("token is required")
	}

	if _, err := uuid.Parse(b.Token); err != nil {
		return errors.New("token is invalid")
	}

	return nil
}

type AuthVerifyResponse struct {
	Status string              `json:"status"`
	Data   models.ResponseUser `json:"data"`
}

func AuthVerify(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body AuthVerifyBody
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			response.HandleClientError(w, err, "invalid request body")
			return
		}

		if err := body.Validate(); err != nil {
			response.HandleClientError(w, err, err.Error())
			return
		}

		token, err := uuid.Parse(body.Token)
		if err != nil {
			response.HandleClientError(w, err, "token is invalid")
			return
		}
		u, err := s.Users.GetByEmailVerificationToken(r.Context(), token)
		if err != nil {
			response.HandleDbError(w, err)
			return
		}

		if u.EmailVerificationTokenExpiresAt != nil && u.EmailVerificationTokenExpiresAt.Before(time.Now()) {
			response.HandleClientError(w, errors.New("token has expired"), "token has expired")
			return
		}

		u, err = s.Users.Verify(r.Context(), u.ID)
		if err != nil {
			response.HandleDbError(w, err)
			return
		}

		res := &AuthVerifyResponse{
			Status: "ok",
			Data:   models.NewResponseUser(u),
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
