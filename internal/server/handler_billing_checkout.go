package server

import (
	"net/http"

	"github.com/urlspace/api/internal/billing"
	"github.com/urlspace/api/internal/user"
)

type responseBillingCheckout struct {
	URL string `json:"url"`
}

type billingCheckoutResponse struct {
	Status string                  `json:"status"`
	Data   responseBillingCheckout `json:"data"`
}

func handleBillingCheckout(userSvc *user.Service, billingSvc *billing.Service) http.HandlerFunc {
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

		url, err := billingSvc.CreateCheckout(r.Context(), billing.CheckoutParams{
			UserID:  u.ID,
			Email:   u.Email,
			IsPro:   u.IsPro,
			IsAdmin: u.IsAdmin,
		})
		if err != nil {
			statusCode, errorMessage := billing.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusOK, billingCheckoutResponse{
			Status: "ok",
			Data: responseBillingCheckout{
				URL: url,
			},
		})
	}
}
