package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/urlspace/api/internal/config"
	"github.com/urlspace/api/internal/user"
)

func userPermissionsMiddleware(svc *user.Service) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := getUserIDFromContext(r.Context())
			if !ok {
				writeJSONError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			u, err := svc.GetById(r.Context(), userID)
			if err != nil {
				if errors.Is(err, user.ErrNotFound) {
					writeJSONError(w, http.StatusUnauthorized, "unauthorized")
					return
				}
				handleServerError(r.Context(), w, err, "failed to look up user")
				return
			}

			ctx := context.WithValue(r.Context(), config.IsProContextKey, u.IsPro)
			ctx = context.WithValue(ctx, config.IsAdminContextKey, u.IsAdmin)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
