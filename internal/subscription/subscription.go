package subscription

import (
	"time"

	"uuid"
)

type Subscription struct {
	ID                  uuid.UUID
	UserID              uuid.UUID
	PolarSubscriptionID string
	Status              string
	CurrentPeriodEnd    time.Time
	CancelAtPeriodEnd   bool
	PolarModifiedAt     time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
