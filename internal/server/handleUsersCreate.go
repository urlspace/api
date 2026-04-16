package server

import (
	"encoding/json"
	"net/http"

	"github.com/hreftools/api/internal/user"
)

type userCreateBody struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	IsAdmin  *bool  `json:"isAdmin"`
	IsPro    *bool  `json:"isPro"`
}

type usersCreateResponse struct {
	Status string            `json:"status"`
	Data   responseUserAdmin `json:"data"`
}

func handleUsersCreate(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body userCreateBody
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			handleClientError(w, err, "invalid request body")
			return
		}

		u, err := svc.AdminCreate(r.Context(), body.Username, body.Email, body.Password, body.IsAdmin, body.IsPro)
		if err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		res := &usersCreateResponse{
			Status: "ok",
			Data:   newResponseUserAdmin(u),
		}

		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
