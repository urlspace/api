package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jumplist/api/internal/db"
	"github.com/jumplist/api/internal/response"
	"github.com/jumplist/api/internal/store"
	"github.com/jumplist/api/internal/utils"
)

type AuthSignupBody struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (b *AuthSignupBody) Validate() error {
	// username
	b.Username = strings.TrimSpace(b.Username)

	if len(b.Username) == 0 {
		return errors.New("username is required")
	}

	if len(b.Username) < store.UserUsernameLengthMin {
		return errors.New("username must be min 3 characters")
	}

	if len(b.Username) > store.UserUsernameLengthMax {
		return errors.New("username must be max 32 characters")
	}

	if b.Username != strings.ToLower(b.Username) {
		return errors.New("username must be lowercase")
	}

	if !regexp.MustCompile(`^[a-z0-9_-]+$`).MatchString(b.Username) {
		return errors.New("username can only contain lowercase characters, numbers, hyphens, and underscores")
	}

	if strings.HasPrefix(b.Username, "-") || strings.HasPrefix(b.Username, "_") {
		return errors.New("username cannot start with hyphen or underscore")
	}

	if strings.HasSuffix(b.Username, "-") || strings.HasSuffix(b.Username, "_") {
		return errors.New("username cannot end with hyphen or underscore")
	}

	if reserved := reservedUsernames[b.Username]; reserved {
		return errors.New("username is reserved")
	}

	// password
	if len(b.Password) == 0 {
		return errors.New("password is required")
	}

	if len(b.Password) < store.UserPasswordLengthMin {
		return errors.New("password must be at least 12 characters")
	}

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

type AuthSignupResponse struct {
	Status string  `json:"status"`
	Data   db.User `json:"data"`
}

func AuthSignup(store *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body AuthSignupBody
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

		passwordHash, err := utils.PasswordHash(body.Password)
		if err != nil {
			response.HandleClientError(w, err, "failed to hash password")
			return
		}
		tokenExpiresAt := time.Now().Add(24 * time.Hour)
		u, err := store.Users.Create(r.Context(), strings.TrimSpace(body.Email), false, uuid.NullUUID{Valid: true, UUID: uuid.New()}, &tokenExpiresAt, passwordHash, strings.TrimSpace(body.Username), false, false)
		if err != nil {
			response.HandleDbError(w, err)
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
