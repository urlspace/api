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
			tokenID, ok := resolveTokenID(r)
			if !ok {
				writeJSONError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			token, err := svc.GetTokenByID(r.Context(), tokenID)
			if err != nil {
				if errors.Is(err, user.ErrNotFound) {
					writeJSONError(w, http.StatusUnauthorized, "unauthorized")
					return
				}
				handleServerError(w, err, "failed to look up token")
				return
			}

			if time.Now().After(token.ExpiresAt) {
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

			// Sliding expiry: renew session tokens that are approaching expiry.
			if token.Type == config.TokenTypeSession && time.Until(token.ExpiresAt) < config.SessionRenewalThreshold {
				go func() {
					// Fire-and-forget renewal. Errors are intentionally swallowed:
					// a failed renewal is non-fatal — the token remains valid for
					// the remainder of its current expiry window and renewal will
					// be retried on the next request.
					_, _ = svc.UpdateTokenExpiresAt(context.Background(), user.TokenUpdateExpiresAtParams{
						ID:        token.ID,
						ExpiresAt: time.Now().Add(config.SessionExpiryDuration),
					})
				}()
			}

			ctx := context.WithValue(r.Context(), config.UserIDContextKey, token.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
