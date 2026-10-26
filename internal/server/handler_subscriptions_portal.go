package server

import (
	"net/http"

	"github.com/urlspace/api/internal/subscription"
)

type responseSubscriptionsPortal struct {
	URL string `json:"url"`
}

type subscriptionsPortalResponse struct {
	Status string                      `json:"status"`
	Data   responseSubscriptionsPortal `json:"data"`
}

func handleSubscriptionsPortal(svc *subscription.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := getUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		url, err := svc.CreatePortal(r.Context(), userID)
		if err != nil {
			statusCode, errorMessage := subscription.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusOK, subscriptionsPortalResponse{
			Status: "ok",
			Data: responseSubscriptionsPortal{
				URL: url,
			},
		})
	}
}
