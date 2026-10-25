package server

import (
	"io"
	"net/http"

	"github.com/urlspace/api/internal/subscription"
)

type subscriptionsWebhookResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func handleSubscriptionsWebhook(subscriptionSvc *subscription.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			handleClientError(r.Context(), w, err, "invalid request body")
			return
		}

		if err := subscriptionSvc.HandleWebhook(r.Context(), r.Header, body); err != nil {
			statusCode, errorMessage := subscription.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusAccepted, subscriptionsWebhookResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}
