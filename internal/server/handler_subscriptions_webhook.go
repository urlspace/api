package server

import (
	"io"
	"net/http"

	"github.com/urlspace/api/internal/subscription"
	"github.com/urlspace/api/internal/uow"
)

type subscriptionsWebhookResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func handleSubscriptionsWebhook(subscriptionSvc *subscription.Service, uowSvc *uow.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			handleClientError(r.Context(), w, err, "invalid request body")
			return
		}

		params, ok, err := subscriptionSvc.ParseWebhook(r.Context(), r.Header, body)
		if err != nil {
			statusCode, errorMessage := subscription.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		if ok {
			if err := uowSvc.SyncSubscription(r.Context(), params); err != nil {
				statusCode, errorMessage := subscription.MapErrorToHTTP(r.Context(), err)
				writeJSONError(w, statusCode, errorMessage)
				return
			}
		}

		writeJSONSuccess(w, http.StatusAccepted, subscriptionsWebhookResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}
