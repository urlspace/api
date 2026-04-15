package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/hreftools/api/internal/user"
	"github.com/hreftools/api/internal/validator"
)

type authResendVerificationBody struct {
	Email string `json:"email"`
}

func (b *authResendVerificationBody) normalize() {
	b.Email = strings.ToLower(strings.TrimSpace(b.Email))
}

func (b *authResendVerificationBody) validate() error {
	if err := validator.Email(b.Email); err != nil {
		return err
	}

	return nil
}

type authResendVerificationResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func handleAuthResendVerification(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body authResendVerificationBody
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

		err := svc.ResendVerification(r.Context(), body.Email)
		if err != nil {
			if errors.Is(err, user.ErrRateLimited) {
				writeJSONError(w, http.StatusTooManyRequests, "verification email already sent, please wait before requesting a new one")
				return
			}
			handleDbError(w, err)
			return
		}

		writeJSONSuccess(w, http.StatusOK, authResendVerificationResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}
