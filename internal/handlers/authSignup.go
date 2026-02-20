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

		email := body.Email
		username := body.Username
		emailVerified := false
		emailVerificationToken := uuid.NullUUID{Valid: true, UUID: uuid.New()}
		emailVerificationTokenExpiresAt := time.Now().Add(time.Hour * 24)
		isAdmin := false
		isPro := false
		u, err := s.Users.Create(r.Context(), email, emailVerified, emailVerificationToken, &emailVerificationTokenExpiresAt, passwordHash, username, isAdmin, isPro)
		if err != nil {
			response.HandleDbError(w, err)
			return
		}

		emailVerifyData := EmailVerifyData{
			Username:  username,
			Email:     email,
			Token:     emailVerificationToken.UUID.String(),
			ExpiresAt: emailVerificationTokenExpiresAt.Format(time.RFC1123),
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
		params := emails.EmailSendParams{
			From:    "href.tools <auth@mail.href.tools>",
			To:      []string{email},
			Text:    bodyText,
			Html:    bodyHtml,
			Subject: "Hello from href.tools",
			ReplyTo: "auth@mail.href.tools",
		}

		err = emailSender.Send(params)
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
