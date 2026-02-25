package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"log"

	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/hreftools/api/internal/emails"
	"github.com/hreftools/api/internal/response"
	"github.com/hreftools/api/internal/store"
	"github.com/hreftools/api/internal/validator"
)

type AuthResendVerificationBody struct {
	Email string `json:"email"`
}

func (b *AuthResendVerificationBody) Normalize() {
	b.Email = strings.ToLower(strings.TrimSpace(b.Email))
}

func (b *AuthResendVerificationBody) Validate() error {
	if err := validator.Email(b.Email); err != nil {
		return err
	}

	return nil
}

type AuthResendVerificationResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
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

		body.Normalize()

		if err := body.Validate(); err != nil {
			response.HandleClientError(w, err, err.Error())
			return
		}

		u, err := s.Users.GetByEmail(r.Context(), body.Email)

		if err != nil {
			// in case there is no user with the email,
			// or the email is already verified,
			// we should return ok
			if errors.Is(err, sql.ErrNoRows) {
				response.WriteJSONSuccess(w, http.StatusOK, &AuthResendVerificationResponse{
					Status: "ok",
					Data:   "ok",
				})
				return
			}
			response.HandleServerError(w, err, "failed to query user")
			return
		}

		// in case the email is verified, we should return ok,
		// to avoid spamming the user with emails
		if u.EmailVerified {
			response.WriteJSONSuccess(w, http.StatusOK, &AuthResendVerificationResponse{
				Status: "ok",
				Data:   "ok",
			})
			return
		}

		// in case that the token has been generated in the last 5 minutes,
		// we should not generate a new token, to avoid spamming the user with emails
		tokenAge := TokenExpiryDuration - time.Until(*u.EmailVerificationTokenExpiresAt)
		if tokenAge < time.Minute*5 {
			log.Println("verification email already sent, please wait before requesting a new one")
			response.WriteJSONError(w, http.StatusTooManyRequests, "verification email already sent, please wait before requesting a new one")
			return
		}

		token := uuid.NullUUID{Valid: true, UUID: uuid.New()}

		templateParams := emails.AuthResendVerificationParams{
			Token: token.UUID.String(),
		}
		bodyHtml, err := emails.RenderTemplateHtml(emails.AuthResendVerificationTemplateHtml, templateParams)
		if err != nil {
			response.HandleServerError(w, err, "failed to render html email template")
			return
		}
		bodyText, err := emails.RenderTemplateTxt(emails.AuthResendVerificationTemplateTxt, templateParams)
		if err != nil {
			response.HandleServerError(w, err, "failed to render text email template")
			return
		}

		params := store.UserUpdateVerificationTokenParams{
			Id:                              u.ID,
			EmailVerificationToken:          token,
			EmailVerificationTokenExpiresAt: new(time.Now().Add(TokenExpiryDuration)),
		}
		u, err = s.Users.UpdateVerificationToken(r.Context(), params)

		if err != nil {
			response.HandleDbError(w, err)
			return
		}

		emailParams := emails.EmailSendParams{
			To:      []string{body.Email},
			Text:    bodyText,
			Html:    bodyHtml,
			Subject: "Verification token has been requested",
		}

		err = emailSender.Send(emailParams)
		if err != nil {
			log.Printf("Failed to send email: %v", err)
		}

		response.WriteJSONSuccess(w, http.StatusOK, &AuthResendVerificationResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}
