package server

import (
	"net/http"

	"github.com/urlspace/api/internal/config"
	"github.com/urlspace/api/internal/user"
)

type authSignoutResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func handleAuthSignout(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID, _ := resolveSessionID(r)

		if err := svc.Signout(r.Context(), sessionID); err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     config.SessionCookieName,
			MaxAge:   -1,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		})

		writeJSONSuccess(w, http.StatusOK, authSignoutResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}
