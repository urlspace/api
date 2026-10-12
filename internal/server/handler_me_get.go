package server

import (
	"net/http"

	"github.com/urlspace/api/internal/user"
)

type responseMe struct {
	responseUser `json:",inline"`
	Stats        responseStats `json:"stats"`
}

type meGetResponse struct {
	Status string     `json:"status"`
	Data   responseMe `json:"data"`
}

func handleMeGet(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := getUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		u, err := svc.GetById(r.Context(), userID)
		if err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		stats, err := svc.GetStats(r.Context(), u)
		if err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusOK, meGetResponse{
			Status: "ok",
			Data: responseMe{
				responseUser: newResponseUser(u),
				Stats:        newResponseStats(stats),
			},
		})
	}
}
