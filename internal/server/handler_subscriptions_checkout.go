package server

import (
	"encoding/json/v2"
	"net/http"

	"github.com/urlspace/api/internal/subscription"
	"github.com/urlspace/api/internal/user"
)

type subscriptionsCheckoutBody struct {
	ReturnTo string `json:"returnTo"`
}

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

		var body subscriptionsCheckoutBody
		if err := json.UnmarshalRead(r.Body, &body, json.RejectUnknownMembers(true)); err != nil {
			handleClientError(r.Context(), w, err, "invalid request body")
			return
		}

		u, err := userSvc.GetById(r.Context(), userID)
		if err != nil {
			statusCode, errorMessage := user.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		url, err := subscriptionSvc.CreateCheckout(r.Context(), subscription.CheckoutParams{
			UserID:      u.ID,
			Email:       u.Email,
			IsUnlimited: u.IsUnlimited,
			IsAdmin:     u.IsAdmin,
			ReturnTo:    body.ReturnTo,
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
