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

type AuthResetPasswordConfirmBody struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

func (b *AuthResetPasswordConfirmBody) Normalize() {
	b.Token = strings.TrimSpace(b.Token)
}

func (b *AuthResetPasswordConfirmBody) Validate() error {
	if err := validator.Token(b.Token); err != nil {
		return err
	}

	if err := validator.Password(b.Password); err != nil {
		return err
	}

	return nil
}

type AuthResetPasswordConfirmResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func AuthResetPasswordConfirm(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body AuthResetPasswordConfirmBody
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

		err := svc.ResetPasswordConfirm(r.Context(), body.Token, body.Password)
		if err != nil {
			if errors.Is(err, user.ErrTokenExpired) {
				response.HandleClientError(w, err, "token has expired")
				return
			}
			response.HandleDbError(w, err)
			return
		}

		response.WriteJSONSuccess(w, http.StatusOK, AuthResetPasswordConfirmResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}
