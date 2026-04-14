package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/hreftools/api/internal/response"
	"github.com/hreftools/api/internal/user"
	"github.com/hreftools/api/internal/validator"
)

type AuthSignupBody struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (b *AuthSignupBody) Normalize() {
	b.Username = strings.ToLower(strings.TrimSpace(b.Username))
	b.Email = strings.ToLower(strings.TrimSpace(b.Email))
}

func (b *AuthSignupBody) Validate() error {
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

type AuthSignupResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func AuthSignup(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body AuthSignupBody
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

		err := svc.Signup(r.Context(), body.Username, body.Email, body.Password)
		if err != nil {
			response.HandleDbError(w, err)
			return
		}

		response.WriteJSONSuccess(w, http.StatusCreated, AuthSignupResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}
