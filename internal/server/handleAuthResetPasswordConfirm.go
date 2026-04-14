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

type authResetPasswordConfirmBody struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

func (b *authResetPasswordConfirmBody) normalize() {
	b.Token = strings.TrimSpace(b.Token)
}

func (b *authResetPasswordConfirmBody) validate() error {
	if err := validator.Token(b.Token); err != nil {
		return err
	}

	if err := validator.Password(b.Password); err != nil {
		return err
	}

	return nil
}

type authResetPasswordConfirmResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func handleAuthResetPasswordConfirm(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body authResetPasswordConfirmBody
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

		err := svc.ResetPasswordConfirm(r.Context(), body.Token, body.Password)
		if err != nil {
			if errors.Is(err, user.ErrTokenExpired) {
				response.HandleClientError(w, err, "token has expired")
				return
			}
			response.HandleDbError(w, err)
			return
		}

		response.WriteJSONSuccess(w, http.StatusOK, authResetPasswordConfirmResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}
