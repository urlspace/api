package collection

import (
	"time"

	"uuid"
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

type PublicCollection struct {
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Author      PublicAuthor
	Links       []PublicLink
}

type PublicAuthor struct {
	DisplayName string
	Username    string
}

type PublicLink struct {
	ID          uuid.UUID
	Title       string
	Description string
	CreatedAt   time.Time
	URL         string
}
