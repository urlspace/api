package server

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/hreftools/api/internal/config"
	"github.com/hreftools/api/internal/response"
	"github.com/hreftools/api/internal/user"
	"github.com/hreftools/api/internal/utils"
)

type middleware func(http.Handler) http.Handler

func middlewareStack(mds ...middleware) middleware {
	return func(next http.Handler) http.Handler {
		for i := len(mds) - 1; i >= 0; i-- {
			next = mds[i](next)
		}
		return next
	}
}

func commonHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		// CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Others
		w.Header().Set("Content-Type", "application/json")

		next.ServeHTTP(w, r)
	})
}

type wrapperWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrapperWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &wrapperWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapped, r)
		log.Println(wrapped.statusCode, r.Method, r.URL.Path, time.Since(start))
	})
}

func maxBodySizeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 64<<10) // 64 KB
		next.ServeHTTP(w, r)
	})
}

func authMiddleware(svc *user.Service) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenID, ok := utils.ResolveTokenID(r)
			if !ok {
				response.WriteJSONError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			token, err := svc.GetTokenByID(r.Context(), tokenID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					response.WriteJSONError(w, http.StatusUnauthorized, "unauthorized")
					return
				}
				response.HandleServerError(w, err, "failed to look up token")
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
				response.WriteJSONError(w, http.StatusUnauthorized, "unauthorized")
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

func adminMiddleware(svc *user.Service) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := utils.UserIDFromContext(r.Context())
			if !ok {
				response.WriteJSONError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			u, err := svc.GetById(r.Context(), userID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					response.WriteJSONError(w, http.StatusUnauthorized, "unauthorized")
					return
				}
				response.HandleServerError(w, err, "failed to look up user")
				return
			}

			if !u.IsAdmin {
				response.WriteJSONError(w, http.StatusForbidden, "forbidden")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
