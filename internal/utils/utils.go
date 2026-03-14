package utils

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/hreftools/api/internal/config"
	"golang.org/x/crypto/bcrypt"
)

func ResolveTokenID(r *http.Request) (uuid.UUID, bool) {
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			if id, err := uuid.Parse(strings.TrimSpace(parts[1])); err == nil {
				return id, true
			}
		}
	}

	if cookie, err := r.Cookie(config.SessionCookieName); err == nil {
		if id, err := uuid.Parse(cookie.Value); err == nil {
			return id, true
		}
	}

	return uuid.UUID{}, false
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(config.UserIDContextKey).(uuid.UUID)
	return id, ok
}

func PasswordHash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func PasswordValidate(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
