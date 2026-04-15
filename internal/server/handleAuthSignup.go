package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/hreftools/api/internal/user"
	"github.com/hreftools/api/internal/validator"
)

type authSignupBody struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (b *authSignupBody) normalize() {
	b.Username = strings.ToLower(strings.TrimSpace(b.Username))
	b.Email = strings.ToLower(strings.TrimSpace(b.Email))
}

func (b *authSignupBody) validate() error {
	if err := validator.Username(b.Username); err != nil {
		return err
	}

	if err := validator.Email(b.Email); err != nil {
		return err
	}

	if err := validator.Password(b.Password); err != nil {
		return err
	}

	return nil
}

type authSignupResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func handleAuthSignup(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body authSignupBody
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

		err := svc.Signup(r.Context(), body.Username, body.Email, body.Password)
		if err != nil {
			handleDbError(w, err)
			return
		}

		writeJSONSuccess(w, http.StatusCreated, authSignupResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}
