package server

import (
	"io"
	"net/http"

	"github.com/urlspace/api/internal/billing"
)

type billingWebhookResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func handleBillingWebhook(billingSvc *billing.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			handleClientError(r.Context(), w, err, "invalid request body")
			return
		}

		if err := billingSvc.HandleWebhook(r.Context(), r.Header, body); err != nil {
			statusCode, errorMessage := billing.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusAccepted, billingWebhookResponse{
			Status: "ok",
			Data:   "ok",
		})
	}
}
