package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const EmailVerificationTokenExpiryDuration = 24 * time.Hour
const PasswordResetTokenExpiryDuration = 1 * time.Hour

const SessionExpiryDuration = 30 * 24 * time.Hour
const SessionRenewalThreshold = 15 * 24 * time.Hour

type contextKey string

const UserIDContextKey contextKey = "userID"

// Decision (future me): __Host- prefix dropped because SSR on url.space can't
// read a cookie host-scoped to api.url.space. Domain set in setSessionCookie.
const SessionCookieName = "session"

type Config struct {
	Port         string
	DatabaseURL  string
	ResendAPIKey string
	AppURL       string
	AdminEmail   string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		Port:         os.Getenv("PORT"),
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		ResendAPIKey: os.Getenv("RESEND_API_KEY"),
		AppURL:       strings.TrimSuffix(os.Getenv("APP_URL"), "/"),
		AdminEmail:   os.Getenv("ADMIN_EMAIL"),
	}

	var missing []string

	if cfg.Port == "" {
		missing = append(missing, "PORT")
	}
	if cfg.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if cfg.ResendAPIKey == "" {
		missing = append(missing, "RESEND_API_KEY")
	}
	if cfg.AppURL == "" {
		missing = append(missing, "APP_URL")
	}
	if cfg.AdminEmail == "" {
		missing = append(missing, "ADMIN_EMAIL")
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}
