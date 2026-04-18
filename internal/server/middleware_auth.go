package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hreftools/api/internal/config"
	"github.com/hreftools/api/internal/user"
)

type AuthConfig struct {
	UseSession bool
	UseToken   bool
}

func authMiddleware(svc *user.Service, cfg AuthConfig) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.UseSession {
				if sessionID, ok := resolveSessionID(r); ok {
					authenticateSession(w, r, svc, sessionID, next)
					return
				}
			}

			if cfg.UseToken {
				if rawToken, ok := resolveBearerToken(r); ok {
					authenticateToken(w, r, svc, rawToken, next)
					return
				}
			}

			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		})
	}
}

func authenticateSession(w http.ResponseWriter, r *http.Request, svc *user.Service, sessionID uuid.UUID, next http.Handler) {
	session, err := svc.GetSessionByID(r.Context(), sessionID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		handleServerError(w, err, "failed to look up session")
		return
	}

	// Clear the cookie on the client to prevent repeated lookups of the same expired session on subsequent requests.
	if time.Now().After(session.ExpiresAt) {
		http.SetCookie(w, &http.Cookie{
			Name:     config.SessionCookieName,
			MaxAge:   -1,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		})
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Sliding expiry: renew sessions that are approaching expiry.
	if time.Until(session.ExpiresAt) < config.SessionRenewalThreshold {
		go func() {
			// Fire-and-forget renewal. Errors are intentionally swallowed:
			// a failed renewal is non-fatal — the session remains valid for
			// the remainder of its current expiry window and renewal will
			// be retried on the next request.
			_, _ = svc.UpdateSessionExpiresAt(context.Background(), user.SessionUpdateExpiresAtParams{
				ID:        session.ID,
				ExpiresAt: time.Now().Add(config.SessionExpiryDuration),
			})
		}()
	}

	ctx := context.WithValue(r.Context(), config.UserIDContextKey, session.UserID)
	next.ServeHTTP(w, r.WithContext(ctx))
}

func authenticateToken(w http.ResponseWriter, r *http.Request, svc *user.Service, rawToken string, next http.Handler) {
	token, err := svc.GetTokenByHash(r.Context(), rawToken)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		handleServerError(w, err, "failed to look up token")
		return
	}

	// Update last used timestamp asynchronously since it's non-critical and we don't want to block the request on a database write.
	go func() {

		// Fire-and-forget update. Errors are intentionally swallowed
		_ = svc.UpdateTokenLastUsedAt(context.Background(), token.ID)
	}()

	ctx := context.WithValue(r.Context(), config.UserIDContextKey, token.UserID)
	next.ServeHTTP(w, r.WithContext(ctx))
}
