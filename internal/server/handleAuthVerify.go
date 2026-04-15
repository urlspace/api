package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/hreftools/api/internal/user"
	"github.com/hreftools/api/internal/validator"
)

type authVerifyBody struct {
	Token string `json:"token"`
}

func (b *authVerifyBody) normalize() {
	b.Token = strings.TrimSpace(b.Token)
}

func (b *authVerifyBody) validate() error {
	if err := validator.Token(b.Token); err != nil {
		return err
	}

	return nil
}

type authVerifyResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func handleAuthVerify(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body authVerifyBody
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			handleClientError(w, err, "invalid request body")
			return
		}

		body.normalize()

		if err := body.validate(); err != nil {
			handleClientError(w, err, err.Error())
			return
		}

		err := svc.Verify(r.Context(), body.Token)
		if err != nil {
			if errors.Is(err, user.ErrTokenExpired) {
				handleClientError(w, err, "token has expired")
				return
			}
			handleDbError(w, err)
			return
		}

		writeJSONSuccess(w, http.StatusOK, authVerifyResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}
