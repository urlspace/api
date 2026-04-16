package server

import (
	"encoding/json"
	"net/http"

	"github.com/hreftools/api/internal/user"
)

type meGetResponse struct {
	Status string       `json:"status"`
	Data   responseUser `json:"data"`
}

func handleMeGet(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := userIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		u, err := svc.GetById(r.Context(), userID)
		if err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		res := &meGetResponse{
			Status: "ok",
			Data:   newResponseUser(u),
		}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
