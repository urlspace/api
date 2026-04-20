package server

import (
	"net/http"

	"github.com/urlspace/api/internal/user"
)

type tokensListResponse struct {
	Status string          `json:"status"`
	Data   []responseToken `json:"data"`
}

func handleTokensList(svc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := userIDFromContext(r.Context())

		list, err := svc.TokenList(r.Context(), userID)
		if err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		items := make([]responseToken, len(list))
		for i, item := range list {
			items[i] = newResponseToken(item)
		}

		writeJSONSuccess(w, http.StatusOK, tokensListResponse{
			Status: "ok",
			Data:   items,
		})
	}
}
