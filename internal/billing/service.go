package billing

import (
	"context"
	"errors"

	"github.com/urlspace/api/internal/polar"
	"uuid"
)

var (
	ErrAlreadyPro = errors.New("already subscribed to pro")
)

type Service struct {
	polarClient *polar.Client
	productID   string
	siteURL     string
}

func NewService(polarClient *polar.Client, productID string, siteURL string) *Service {
	return &Service{
		polarClient: polarClient,
		productID:   productID,
		siteURL:     siteURL,
	}
}

type CheckoutParams struct {
	UserID uuid.UUID
	Email  string
	IsPro  bool
}

func (s *Service) CreateCheckout(ctx context.Context, params CheckoutParams) (string, error) {
	if params.IsPro {
		return "", ErrAlreadyPro
	}

	checkout, err := s.polarClient.CreateCheckout(ctx, polar.CheckoutCreateParams{
		Products:           []string{s.productID},
		ExternalCustomerID: params.UserID.String(),
		CustomerEmail:      params.Email,
		SuccessURL:         s.siteURL + "/settings?checkout_id={CHECKOUT_ID}",
		ReturnURL:          s.siteURL + "/pricing",
	})
	if err != nil {
		return "", err
	}

	return checkout.URL, nil
}
