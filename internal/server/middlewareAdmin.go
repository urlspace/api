package server

import (
	"errors"
	"net/http"

	"github.com/hreftools/api/internal/user"
	"github.com/hreftools/api/internal/utils"
)

func adminMiddleware(svc *user.Service) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := utils.UserIDFromContext(r.Context())
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
				handleServerError(w, err, "failed to look up user")
				return
			}

			if !u.IsAdmin {
				writeJSONError(w, http.StatusForbidden, "forbidden")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
