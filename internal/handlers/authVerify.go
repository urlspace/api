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
	"github.com/hreftools/api/internal/validator"
)

type AuthVerifyBody struct {
	Token string `json:"token"`
}

func (b *AuthVerifyBody) Normalize() {
	b.Token = strings.TrimSpace(b.Token)
}

func (b *AuthVerifyBody) Validate() error {
	if err := validator.Token(b.Token); err != nil {
		return err
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

		body.Normalize()

		if err := body.Validate(); err != nil {
			response.HandleClientError(w, err, err.Error())
			return
		}

		// im swallowing the error here as the token is already
		// validated as a uuid, so it should never fail to parse
		token, _ := uuid.Parse(body.Token)

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

		res := AuthVerifyResponse{
			Status: "ok",
			Data:   models.NewResponseUser(u),
		}

		response.WriteJSONSuccess(w, http.StatusOK, res)
	}
}
