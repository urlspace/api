package collection

import (
	"time"

	"github.com/google/uuid"
)

type Collection struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Name        string
	Description string
	Public      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CollectionWithLinkCount struct {
	Collection
	LinkCount int
}
