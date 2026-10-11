package server

import (
	"encoding/json/v2"
	"net/http"
	"strconv"

	"github.com/mileusna/useragent"
	"github.com/urlspace/api/internal/user"
)

type authSigninBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authSigninResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func handleAuthSignin(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body authSigninBody
		if err := json.UnmarshalRead(r.Body, &body, json.RejectUnknownMembers(true)); err != nil {
			handleClientError(r.Context(), w, err, "invalid request body")
			return
		}

		// No good reason for 1024 other than I like it.
		// Observe registered UA lengths before revisiting this limit.
		const maxUserAgentBytes = 1024
		ua := r.Header.Get("User-Agent")
		if len(ua) > maxUserAgentBytes {
			ua = ua[:maxUserAgentBytes]
		}
		description := sessionDescription(ua)
		result, err := svc.Signin(r.Context(), body.Email, body.Password, description)
		if err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		setSessionCookie(w, r, result.Session, result.ExpiresAt)

		writeJSONSuccess(w, http.StatusOK, authSigninResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}

func sessionDescription(ua string) string {
	parsed := useragent.Parse(ua)

	browser := parsed.Name
	if browser != "" && parsed.VersionNo.Major > 0 {
		browser += " " + strconv.Itoa(parsed.VersionNo.Major)
	}

	switch {
	case browser != "" && parsed.OS != "":
		return browser + " on " + parsed.OS
	case browser != "":
		return browser
	case parsed.OS != "":
		return parsed.OS
	default:
		return "Unknown device"
	}
}
