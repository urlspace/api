package server

import (
	"net/http"

	"github.com/urlspace/api/internal/subscription"
)

type subscriptionsListResponse struct {
	Status string                 `json:"status"`
	Data   []responseSubscription `json:"data"`
}

func handleSubscriptionsList(svc *subscription.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := getUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		list, err := svc.List(r.Context(), userID)
		if err != nil {
			statusCode, errorMessage := subscription.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		items := make([]responseSubscription, len(list))
		for i, item := range list {
			items[i] = newResponseSubscription(item)
		}

		writeJSONSuccess(w, http.StatusOK, subscriptionsListResponse{
			Status: "ok",
			Data:   items,
		})
	}
}
