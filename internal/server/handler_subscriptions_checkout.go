package server

import (
	"net/http"

	"github.com/urlspace/api/internal/subscription"
	"github.com/urlspace/api/internal/user"
)

type responseSubscriptionsCheckout struct {
	URL string `json:"url"`
}

type subscriptionsCheckoutResponse struct {
	Status string                        `json:"status"`
	Data   responseSubscriptionsCheckout `json:"data"`
}

func handleSubscriptionsCheckout(userSvc *user.Service, subscriptionSvc *subscription.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := getUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		u, err := userSvc.GetById(r.Context(), userID)
		if err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		url, err := subscriptionSvc.CreateCheckout(r.Context(), subscription.CheckoutParams{
			UserID:  u.ID,
			Email:   u.Email,
			IsPro:   u.IsPro,
			IsAdmin: u.IsAdmin,
		})
		if err != nil {
			statusCode, errorMessage := subscription.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusOK, subscriptionsCheckoutResponse{
			Status: "ok",
			Data: responseSubscriptionsCheckout{
				URL: url,
			},
		})
	}
}
