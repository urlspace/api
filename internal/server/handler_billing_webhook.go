package server

import (
	"net/http"
)

type billingWebhookResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func handleBillingWebhook() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSONSuccess(w, http.StatusOK, billingWebhookResponse{
			Status: "ok",
			Data:   "hello world from webhook",
		})
	}
}
