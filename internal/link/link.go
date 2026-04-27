package link

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Link struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	Title           string
	Description     string
	URL             string
	CollectionID    *uuid.UUID
	CollectionTitle string // populated by Get/List via JOIN, empty on Create/Update/Delete
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

var (
	// validation title
	ErrValidationTitleLength            = errors.New("title must be between 3 and 255 characters")
	ErrValidationTitleInvalidCharacters = errors.New("title must not contain control characters")

	// validation description
	ErrValidationDescriptionLength            = errors.New("description must be less than 512 characters")
	ErrValidationDescriptionInvalidCharacters = errors.New("description must not contain control characters")

	// validation url
	ErrValidationURLFormat  = errors.New("url is invalid")
	ErrValidationURLTooLong = errors.New("url must be at most 2048 characters")
	ErrValidationURLPrivate = errors.New("url must not point to a private or local address")

	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

type CreateParams struct {
	UserID       uuid.UUID
	Title        string
	Description  string
	URL          string
	CollectionID *uuid.UUID
}

type UpdateParams struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	Title        string
	Description  string
	URL          string
	CollectionID *uuid.UUID
}

type Repository interface {
	List(ctx context.Context, userID uuid.UUID) ([]Link, error)
	Get(ctx context.Context, id uuid.UUID, userID uuid.UUID) (Link, error)
	Create(ctx context.Context, params CreateParams) (Link, error)
	Update(ctx context.Context, params UpdateParams) (Link, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) (Link, error)
}
