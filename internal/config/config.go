package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const EmailVerificationTokenExpiryDuration = 24 * time.Hour
const PasswordResetTokenExpiryDuration = 1 * time.Hour
const EmailChangeCodeExpiryDuration = 15 * time.Minute

const SessionExpiryDuration = 30 * 24 * time.Hour
const SessionRenewalThreshold = 15 * 24 * time.Hour

const FreeLinksLimit = 1000
const FreeCollectionsLimit = 100
const FreeTagsLimit = 100

type contextKey string

const UserIDContextKey contextKey = "userID"
const SessionIDContextKey contextKey = "sessionID"
const IsProContextKey contextKey = "isPro"
const IsAdminContextKey contextKey = "isAdmin"

// Decision (future me): __Host- prefix dropped because SSR on url.space can't
// read a cookie host-scoped to api.url.space. Domain set in setSessionCookie.
const SessionCookieName = "session"

type Config struct {
	Port         string
	DatabaseURL  string
	ResendAPIKey string
	SiteURL      string
	DashboardURL string
	AdminEmail   string

	PolarAPIURL        string
	PolarAccessToken   string
	PolarProductID     string
	PolarWebhookSecret string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		Port:         os.Getenv("PORT"),
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		ResendAPIKey: os.Getenv("RESEND_API_KEY"),
		SiteURL:      strings.TrimSuffix(os.Getenv("SITE_URL"), "/"),
		DashboardURL: strings.TrimSuffix(os.Getenv("DASHBOARD_URL"), "/"),
		AdminEmail:   os.Getenv("ADMIN_EMAIL"),

		PolarAPIURL:        strings.TrimSuffix(os.Getenv("POLAR_API_URL"), "/"),
		PolarAccessToken:   os.Getenv("POLAR_ACCESS_TOKEN"),
		PolarProductID:     os.Getenv("POLAR_PRODUCT_ID"),
		PolarWebhookSecret: os.Getenv("POLAR_WEBHOOK_SECRET"),
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
	if cfg.SiteURL == "" {
		missing = append(missing, "SITE_URL")
	}
	if cfg.DashboardURL == "" {
		missing = append(missing, "DASHBOARD_URL")
	}
	if cfg.AdminEmail == "" {
		missing = append(missing, "ADMIN_EMAIL")
	}
	if cfg.PolarAPIURL == "" {
		missing = append(missing, "POLAR_API_URL")
	}
	if cfg.PolarAccessToken == "" {
		missing = append(missing, "POLAR_ACCESS_TOKEN")
	}
	if cfg.PolarProductID == "" {
		missing = append(missing, "POLAR_PRODUCT_ID")
	}
	if cfg.PolarWebhookSecret == "" {
		missing = append(missing, "POLAR_WEBHOOK_SECRET")
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}
