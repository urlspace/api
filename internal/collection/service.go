package collection

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	// Name validation errors.
	ErrValidationNameLength            = errors.New("name must be between 2 and 255 characters")
	ErrValidationNameInvalidCharacters = errors.New("name must not contain control characters")

	// Description validation errors.
	ErrValidationDescriptionLength            = errors.New("description must be less than 512 characters")
	ErrValidationDescriptionInvalidCharacters = errors.New("description must not contain control characters")

	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

type CreateParams struct {
	UserID      uuid.UUID
	Name        string
	Description string
	Public      bool
}

type UpdateParams struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Name        string
	Description string
	Public      bool
}

type Repository interface {
	List(ctx context.Context, userID uuid.UUID) ([]CollectionWithLinkCount, error)
	Get(ctx context.Context, id uuid.UUID, userID uuid.UUID) (Collection, error)
	Create(ctx context.Context, params CreateParams) (Collection, error)
	Update(ctx context.Context, params UpdateParams) (Collection, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) (Collection, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]CollectionWithLinkCount, error) {
	return s.repo.List(ctx, userID)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID, userID uuid.UUID) (Collection, error) {
	return s.repo.Get(ctx, id, userID)
}

func (s *Service) Create(ctx context.Context, params CreateParams) (Collection, error) {
	name, err := ValidateName(params.Name)
	if err != nil {
		return Collection{}, err
	}
	description, err := ValidateDescription(params.Description)
	if err != nil {
		return Collection{}, err
	}

	return s.repo.Create(ctx, CreateParams{
		UserID:      params.UserID,
		Name:        name,
		Description: description,
		Public:      params.Public,
	})
}

func (s *Service) Update(ctx context.Context, params UpdateParams) (Collection, error) {
	name, err := ValidateName(params.Name)
	if err != nil {
		return Collection{}, err
	}
	description, err := ValidateDescription(params.Description)
	if err != nil {
		return Collection{}, err
	}

	return s.repo.Update(ctx, UpdateParams{
		ID:          params.ID,
		UserID:      params.UserID,
		Name:        name,
		Description: description,
		Public:      params.Public,
	})
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) (Collection, error) {
	return s.repo.Delete(ctx, id, userID)
}
