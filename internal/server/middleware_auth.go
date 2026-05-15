package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/urlspace/api/internal/config"
	"github.com/urlspace/api/internal/user"
)

type AuthConfig struct {
	UseSession bool
	UseToken   bool
}

func authMiddleware(svc *user.Service, cfg AuthConfig) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.UseSession {
				if session, ok := resolveSession(r); ok {
					authenticateSession(w, r, svc, session, next)
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

func authenticateSession(w http.ResponseWriter, r *http.Request, svc *user.Service, session string, next http.Handler) {
	sess, err := svc.GetSession(r.Context(), session)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		handleServerError(r.Context(), w, err, "failed to look up session")
		return
	}

	// Decision (future me): session is past its expiry. Reject the request,
	// clear the cookie on the client, and fire-and-forget a delete so the
	// dead row doesn't linger in the table. Errors are intentionally swallowed:
	// a failed delete just means the row will be cleaned up on the next
	// expired-session hit, or not at all if the user never returns — neither
	// is a correctness problem because the middleware refuses the session
	// regardless. WithoutCancel preserves trace context across the detached
	// call. Recover protects the process from a panic in pgx/tracing/etc.,
	// since an unrecovered panic in a goroutine crashes the whole API server.
	if time.Now().After(sess.ExpiresAt) {
		clearSessionCookie(w)
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		detached := context.WithoutCancel(r.Context())
		go func() {
			defer func() {
				if rec := recover(); rec != nil {
					slog.ErrorContext(detached, "expired session cleanup panicked", "recover", rec)
				}
			}()
			_ = svc.Signout(detached, session)
		}()
		return
	}

	// Decision (future me): sliding expiry — refresh both the cookie and the
	// DB row when the session is approaching expiry, so an actively-used
	// session never dies. The cookie value itself stays the same; we're
	// extending the lifetime of an existing credential, not rotating it.
	// Fire-and-forget renewal. Errors are intentionally swallowed: a failed
	// renewal is non-fatal — the session remains valid for the remainder of
	// its current expiry window and renewal will be retried on the next
	// request. WithoutCancel preserves trace context across the detached
	// call. Recover protects the process from a panic in pgx/tracing/etc.,
	// since an unrecovered panic in a goroutine crashes the whole API server.
	if time.Until(sess.ExpiresAt) < config.SessionRenewalThreshold {
		newExpiresAt := time.Now().Add(config.SessionExpiryDuration)
		setSessionCookie(w, session, newExpiresAt)
		detached := context.WithoutCancel(r.Context())
		go func() {
			defer func() {
				if rec := recover(); rec != nil {
					slog.ErrorContext(detached, "session renewal panicked", "recover", rec)
				}
			}()
			_, _ = svc.UpdateSessionExpiresAt(detached, user.SessionUpdateExpiresAtParams{
				ID:        sess.ID,
				ExpiresAt: newExpiresAt,
			})
		}()
	}

	ctx := context.WithValue(r.Context(), config.UserIDContextKey, sess.UserID)
	next.ServeHTTP(w, r.WithContext(ctx))
}

func authenticateToken(w http.ResponseWriter, r *http.Request, svc *user.Service, rawToken string, next http.Handler) {
	token, err := svc.GetTokenByHash(r.Context(), rawToken)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		handleServerError(r.Context(), w, err, "failed to look up token")
		return
	}

	// Decision (future me): update last_used_at asynchronously since it's
	// non-critical and we don't want to block the request on a database
	// write. Errors are intentionally swallowed — a missed timestamp update
	// has no security or correctness consequence. WithoutCancel preserves
	// trace context across the detached call. Recover protects the process
	// from a panic in pgx/tracing/etc., since an unrecovered panic in a
	// goroutine crashes the whole API server.
	detached := context.WithoutCancel(r.Context())
	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				slog.ErrorContext(detached, "token last_used_at update panicked", "recover", rec)
			}
		}()
		_ = svc.UpdateTokenLastUsedAt(detached, token.ID)
	}()

	ctx := context.WithValue(r.Context(), config.UserIDContextKey, token.UserID)
	next.ServeHTTP(w, r.WithContext(ctx))
}
