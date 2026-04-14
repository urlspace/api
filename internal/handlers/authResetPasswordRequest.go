package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/hreftools/api/internal/response"
	"github.com/hreftools/api/internal/user"
	"github.com/hreftools/api/internal/validator"
)

type AuthResetPasswordRequestBody struct {
	Email string `json:"email"`
}

func (b *AuthResetPasswordRequestBody) Normalize() {
	b.Email = strings.ToLower(strings.TrimSpace(b.Email))
}

func (b *AuthResetPasswordRequestBody) Validate() error {
	if err := validator.Email(b.Email); err != nil {
		return err
	}

	return nil
}

type AuthResetPasswordRequestResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func AuthResetPasswordRequest(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body AuthResetPasswordRequestBody
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

		err := svc.ResetPasswordRequest(r.Context(), body.Email)
		if err != nil {
			if errors.Is(err, user.ErrRateLimited) {
				response.WriteJSONError(w, http.StatusTooManyRequests, "password reset email already sent, please wait before requesting a new one")
				return
			}
			response.HandleDbError(w, err)
			return
		}

		response.WriteJSONSuccess(w, http.StatusOK, AuthResetPasswordRequestResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}
