package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/hreftools/api/internal/response"
	"github.com/hreftools/api/internal/user"
	"github.com/hreftools/api/internal/validator"
)

type userCreateBody struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	IsAdmin  *bool  `json:"isAdmin"`
	IsPro    *bool  `json:"isPro"`
}

func (b *userCreateBody) normalize() {
	b.Username = strings.ToLower(strings.TrimSpace(b.Username))
	b.Email = strings.ToLower(strings.TrimSpace(b.Email))
}

func (b *userCreateBody) validate() error {
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

type usersCreateResponse struct {
	Status string            `json:"status"`
	Data   responseUserAdmin `json:"data"`
}

func handleUsersCreate(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body userCreateBody
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			response.HandleClientError(w, err, "invalid request body")
			return
		}

		body.normalize()

		if err := body.validate(); err != nil {
			response.HandleClientError(w, err, err.Error())
			return
		}

		u, err := svc.AdminCreate(r.Context(), body.Username, body.Email, body.Password, *body.IsAdmin, *body.IsPro)
		if err != nil {
			response.HandleDbError(w, err)
			return
		}

		res := &usersCreateResponse{
			Status: "ok",
			Data:   newResponseUserAdmin(u),
		}

		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
