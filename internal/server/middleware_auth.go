package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/hreftools/api/internal/config"
	"github.com/hreftools/api/internal/user"
)

func authMiddleware(svc *user.Service) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionID, ok := resolveSessionID(r)
			if !ok {
				writeJSONError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			session, err := svc.GetSessionByID(r.Context(), sessionID)
			if err != nil {
				if errors.Is(err, user.ErrNotFound) {
					writeJSONError(w, http.StatusUnauthorized, "unauthorized")
					return
				}
				handleServerError(w, err, "failed to look up session")
				return
			}

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
		})
	}
}
