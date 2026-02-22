package handlers

import (
	"encoding/json"
	"errors"

	// "fmt"
	"net/http"
	"net/mail"
	"strings"

	// "time"

	// "github.com/google/uuid"
	"github.com/hreftools/api/internal/db"
	"github.com/hreftools/api/internal/emails"
	"github.com/hreftools/api/internal/response"
	"github.com/hreftools/api/internal/store"
)

type AuthResendVerificationBody struct {
	Email string `json:"email"`
}

func (b *AuthResendVerificationBody) Validate() error {
	// email
	b.Email = strings.TrimSpace(b.Email)

	if len(b.Email) == 0 {
		return errors.New("email is required")
	}

	// validate format RFC 5322
	if _, err := mail.ParseAddress(b.Email); err != nil {
		return errors.New("email format is invalid")
	}

	// limit length as pet smtp spec RFC 5321
	if len(b.Email) > 254 {
		return errors.New("email must be at most 254 characters")
	}

	return nil
}

type AuthResendVerificationResponse struct {
	Status string  `json:"status"`
	Data   db.User `json:"data"`
}

func AuthResendVerification(s *store.Store, emailSender emails.EmailSender) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body AuthResendVerificationBody
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			response.HandleClientError(w, err, "invalid request body")
			return
		}

		if err := body.Validate(); err != nil {
			response.HandleClientError(w, err, err.Error())
			return
		}

		email := strings.TrimSpace(body.Email)

		u, err := s.Users.GetByEmail(r.Context(), email)
		if err != nil {
			response.HandleDbError(w, err)
			return
		}

		// if
		// emailVerificationToken := uuid.New()
		// emailVerificationTokenExpiresAt := time.Now().Add(24 * time.Hour)
		// u, err := store.Users.UpdateVerificationToken(r.Context(), email, uuid.NullUUID{Valid: true, UUID: emailVerificationToken}, &emailVerificationTokenExpiresAt)
		if err != nil {
			response.HandleDbError(w, err)
			return
		}

		params := emails.EmailSendParams{
			To: []string{email},
			// Text:    fmt.Sprintf("Welcome to href.tools!\n\nYour username: %s\nYour email: %s\n\nPlease verify your email using the following token: %s\nThis token will expire on %s.\n\nThank you for joining href.tools!", u.Username, email, emailVerificationToken.String(), emailVerificationTokenExpiresAt.Format(time.RFC1123)),
			Subject: "Hello from href.tools",
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
