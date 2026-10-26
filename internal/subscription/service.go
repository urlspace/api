package subscription

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/urlspace/api/internal/polar"
	"uuid"
)

var (
	ErrAlreadyPro   = errors.New("already subscribed to pro")
	ErrAdmin        = errors.New("admins already have pro features")
	ErrUserNotFound = errors.New("user not found")
)

type Repository interface {
	Upsert(ctx context.Context, params UpsertParams) (bool, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]Subscription, error)
}

type UpsertParams struct {
	UserID              uuid.UUID
	PolarSubscriptionID string
	Status              string
	CurrentPeriodEnd    time.Time
	CancelAtPeriodEnd   bool
	PolarModifiedAt     time.Time
}

type Service struct {
	repo        Repository
	polarClient *polar.Client
	productID   string
	siteURL     string
}

func NewService(repo Repository, polarClient *polar.Client, productID string, siteURL string) *Service {
	return &Service{
		repo:        repo,
		polarClient: polarClient,
		productID:   productID,
		siteURL:     siteURL,
	}
}

func IsProStatus(status string) bool {
	switch status {
	case "active", "trialing", "past_due":
		return true
	}
	return false
}

type CheckoutParams struct {
	UserID  uuid.UUID
	Email   string
	IsPro   bool
	IsAdmin bool
}

func (s *Service) CreateCheckout(ctx context.Context, params CheckoutParams) (string, error) {
	if params.IsAdmin {
		return "", ErrAdmin
	}
	if params.IsPro {
		return "", ErrAlreadyPro
	}

	customer, err := s.polarClient.GetCustomerByExternalID(ctx, params.UserID.String())
	if errors.Is(err, polar.ErrCustomerNotFound) {
		customer, err = s.polarClient.CreateCustomer(ctx, polar.CustomerCreateParams{
			Email:      params.Email,
			ExternalID: params.UserID.String(),
		})
	}
	if err != nil {
		return "", err
	}

	checkout, err := s.polarClient.CreateCheckout(ctx, polar.CheckoutCreateParams{
		Products:   []string{s.productID},
		CustomerID: customer.ID,
		SuccessURL: s.siteURL + "/settings?checkout_id={CHECKOUT_ID}",
		ReturnURL:  s.siteURL + "/pricing",
	})
	if err != nil {
		return "", err
	}

	return checkout.URL, nil
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]Subscription, error) {
	return s.repo.ListByUserID(ctx, userID)
}

func (s *Service) CreatePortal(ctx context.Context, userID uuid.UUID) (string, error) {
	session, err := s.polarClient.CreateCustomerSession(ctx, polar.CustomerSessionCreateParams{
		ExternalCustomerID: userID.String(),
		ReturnURL:          s.siteURL + "/settings",
	})
	if err != nil {
		return "", err
	}

	return session.CustomerPortalURL, nil
}

// ParseWebhook verifies and decodes a Polar webhook. It returns one of:
//   - (params, true, nil): a valid subscription event for a known user;
//     params should be synced.
//   - (UpsertParams{}, false, nil): a valid webhook with nothing to sync
//     (an unhandled event type, or a missing or invalid external_id); the
//     caller acknowledges it so Polar doesn't retry.
//   - (UpsertParams{}, false, err): verification or decoding failed.
func (s *Service) ParseWebhook(ctx context.Context, header http.Header, body []byte) (UpsertParams, bool, error) {
	event, err := s.polarClient.ParseWebhook(header, body)
	if err != nil {
		return UpsertParams{}, false, err
	}

	slog.InfoContext(ctx, "polar webhook received", slog.String("type", event.Type), slog.String("webhook_id", header.Get("webhook-id")))

	switch event.Type {
	case "subscription.created", "subscription.updated", "subscription.uncanceled", "subscription.past_due", "subscription.revoked":
	default:
		return UpsertParams{}, false, nil
	}

	sub, err := event.Subscription()
	if err != nil {
		return UpsertParams{}, false, err
	}

	if sub.Customer.ExternalID == nil {
		slog.WarnContext(ctx, "polar subscription without external customer id", slog.String("polar_subscription_id", sub.ID))
		return UpsertParams{}, false, nil
	}
	userID, err := uuid.Parse(*sub.Customer.ExternalID)
	if err != nil {
		slog.WarnContext(ctx, "polar subscription with invalid external customer id", slog.String("polar_subscription_id", sub.ID), slog.String("external_id", *sub.Customer.ExternalID))
		return UpsertParams{}, false, nil
	}

	modifiedAt := sub.CreatedAt
	if sub.ModifiedAt != nil {
		modifiedAt = *sub.ModifiedAt
	}

	return UpsertParams{
		UserID:              userID,
		PolarSubscriptionID: sub.ID,
		Status:              sub.Status,
		CurrentPeriodEnd:    sub.CurrentPeriodEnd,
		CancelAtPeriodEnd:   sub.CancelAtPeriodEnd,
		PolarModifiedAt:     modifiedAt,
	}, true, nil
}
