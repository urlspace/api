package server

import (
	"encoding/json"
	"net/http"

	"github.com/hreftools/api/internal/config"
	"github.com/hreftools/api/internal/user"
)

type authDeleteBody struct {
	Password string `json:"password"`
}

type authDeleteResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func handleAuthDelete(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := userIDFromContext(r.Context())

		var body authDeleteBody
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			handleClientError(w, err, "invalid request body")
			return
		}

		if err := svc.DeleteSelf(r.Context(), userID, body.Password); err != nil {
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

		writeJSONSuccess(w, http.StatusOK, authDeleteResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}
