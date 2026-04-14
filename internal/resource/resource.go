package resource

import (
	"time"

	"github.com/google/uuid"
)

type Resource struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Title       string
	Description string
	Url         string
	Favourite   bool
	ReadLater   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
