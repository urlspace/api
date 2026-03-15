package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/hreftools/api/internal/models"
	"github.com/hreftools/api/internal/response"
	"github.com/hreftools/api/internal/store"
	"github.com/hreftools/api/internal/utils"
	"github.com/hreftools/api/internal/validator"
)

type UserCreateBody struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	IsAdmin  *bool  `json:"isAdmin"`
	IsPro    *bool  `json:"isPro"`
}

func (b *UserCreateBody) Normalize() {
	b.Username = strings.ToLower(strings.TrimSpace(b.Username))
	b.Email = strings.ToLower(strings.TrimSpace(b.Email))
}

func (b *UserCreateBody) Validate() error {
	if err := validator.Username(b.Username); err != nil {
		return err
	}

	if err := validator.Email(b.Email); err != nil {
		return err
	}

	if err := validator.Password(b.Password); err != nil {
		return err
	}

	if b.IsAdmin == nil {
		return errors.New("isAdmin field is required")
	}

	if b.IsPro == nil {
		return errors.New("isPro field is required")
	}

	return nil
}

type UserCreateResponse struct {
	Status string                   `json:"status"`
	Data   models.ResponseUserAdmin `json:"data"`
}

func UserCreate(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body UserCreateBody
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

		passwordHash, err := utils.PasswordHash(body.Password)
		if err != nil {
			response.HandleClientError(w, err, "failed to hash password")
			return
		}
		params := store.UserCreateParams{
			Email:                           body.Email,
			EmailVerified:                   true,
			EmailVerificationToken:          uuid.NullUUID{},
			EmailVerificationTokenExpiresAt: nil,
			Password:                        passwordHash,
			Username:                        body.Username,
			IsAdmin:                         *body.IsAdmin,
			IsPro:                           *body.IsPro,
		}
		u, err := s.Users.Create(r.Context(), params)
		if err != nil {
			response.HandleDbError(w, err)
			return
		}

		response := &UserCreateResponse{
			Status: "ok",
			Data:   models.NewResponseUserAdmin(u),
		}

		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
