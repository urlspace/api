package server

import (
	"net/http"

	"github.com/hreftools/api/internal/config"
	"github.com/hreftools/api/internal/response"
	"github.com/hreftools/api/internal/user"
	"github.com/hreftools/api/internal/utils"
)

type authSignoutResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func handleAuthSignout(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenID, _ := utils.ResolveTokenID(r)

		if err := svc.Signout(r.Context(), tokenID); err != nil {
			response.HandleServerError(w, err, "failed to delete session")
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

		response.WriteJSONSuccess(w, http.StatusOK, authSignoutResponse{
			Status: "ok",
			Data:   "signed out",
		})
	}
}
