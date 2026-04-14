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
	Status string `json:"status"`
	Data   string `json:"data"`
}

func AuthVerify(svc *user.Service) http.HandlerFunc {
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

		err := svc.Verify(r.Context(), body.Token)
		if err != nil {
			if errors.Is(err, user.ErrTokenExpired) {
				response.HandleClientError(w, err, "token has expired")
				return
			}
			response.HandleDbError(w, err)
			return
		}

		response.WriteJSONSuccess(w, http.StatusOK, AuthVerifyResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}
