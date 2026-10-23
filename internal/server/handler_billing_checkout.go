package server

import (
	"net/http"
)

type billingCheckoutResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func handleBillingCheckout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSONSuccess(w, http.StatusOK, billingCheckoutResponse{
			Status: "ok",
			Data:   "hello world from checkout",
		})
	}
}
