package server

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/hreftools/api/internal/response"
	"github.com/hreftools/api/internal/user"
	"github.com/hreftools/api/internal/utils"
)

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
