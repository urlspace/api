package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hreftools/api/internal/db"
	"github.com/hreftools/api/internal/emails"
	"github.com/hreftools/api/internal/response"
	"github.com/hreftools/api/internal/store"
	"github.com/hreftools/api/internal/utils"
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
	Status string  `json:"status"`
	Data   db.User `json:"data"`
}

func AuthSignup(s *store.Store, emailSender emails.EmailSender) http.HandlerFunc {
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

		passwordHash, err := utils.PasswordHash(body.Password)
		if err != nil {
			response.HandleClientError(w, err, "failed to hash password")
			return
		}

		token := uuid.NullUUID{Valid: true, UUID: uuid.New()}
		params := store.UserCreateParams{
			Email:                           body.Email,
			EmailVerified:                   false,
			EmailVerificationToken:          token,
			EmailVerificationTokenExpiresAt: new(time.Now().Add(time.Hour * 24)),
			Password:                        passwordHash,
			Username:                        body.Username,
			IsAdmin:                         false,
			IsPro:                           false,
		}
		u, err := s.Users.Create(r.Context(), params)
		if err != nil {
			response.HandleDbError(w, err)
			return
		}

		emailVerifyData := EmailVerifyData{
			Username: body.Username,
			Email:    body.Email,
			Token:    token.UUID.String(),
		}
		bodyHtml, err := emailVerifyRenderHtml(emailVerifyData)
		if err != nil {
			response.HandleClientError(w, err, "failed to render html email template")
			return
		}
		bodyText, err := emailVerifyRenderTxt(emailVerifyData)
		if err != nil {
			response.HandleClientError(w, err, "failed to render text email template")
			return
		}
		emailParams := emails.EmailSendParams{
			From:    "href.tools <auth@mail.href.tools>",
			To:      []string{body.Email},
			Text:    bodyText,
			Html:    bodyHtml,
			Subject: "Hello from href.tools",
			ReplyTo: "mail@href.tools",
		}

		err = emailSender.Send(emailParams)
		if err != nil {
			response.HandleClientError(w, err, err.Error())
			return
		}

		response := &AuthSignupResponse{
			Status: "ok",
			Data:   u,
		}

		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
