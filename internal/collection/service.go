package collection

import (
	"context"
	"errors"

	"uuid"
)

var (
	// Name validation errors.
	ErrValidationNameLength            = errors.New("name must be between 2 and 128 characters")
	ErrValidationNameInvalidCharacters = errors.New("name must not contain control characters")

	// Username validation errors, copied from the user package (see
	// validateUsername).
	ErrValidationUsernameRequired   = errors.New("username is required")
	ErrValidationUsernameTooShort   = errors.New("username must be min 3 characters")
	ErrValidationUsernameTooLong    = errors.New("username must be max 32 characters")
	ErrValidationUsernameCharacters = errors.New("username can only contain lowercase characters, numbers, hyphens, and underscores")
	ErrValidationUsernamePrefix     = errors.New("username cannot start with hyphen or underscore")
	ErrValidationUsernameSuffix     = errors.New("username cannot end with hyphen or underscore")
	ErrValidationUsernameReserved   = errors.New("username is reserved")

	// Slug validation errors.
	ErrValidationSlugLength            = errors.New("slug must be between 2 and 128 characters")
	ErrValidationSlugInvalidCharacters = errors.New("slug must contain only lowercase letters, digits and single hyphens")

	// Description validation errors.
	ErrValidationDescriptionLength            = errors.New("description must be at most 1024 characters")
	ErrValidationDescriptionInvalidCharacters = errors.New("description must not contain control characters")

	ErrNotFound           = errors.New("not found")
	ErrConflict           = errors.New("conflict")
	ErrSlugConflict       = errors.New("slug is already taken")
	ErrCloneOwnCollection = errors.New("you cannot clone your own collection")
)

type CreateParams struct {
	UserID      uuid.UUID
	Name        string
	Slug        string
	Description string
	Public      bool
}

type UpdateParams struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Name        string
	Slug        string
	Description string
	Public      bool
}

type Repository interface {
	ClonePublic(ctx context.Context, sourceID uuid.UUID, userID uuid.UUID) (Collection, error)
	List(ctx context.Context, userID uuid.UUID) ([]CollectionWithLinkCount, error)
	Get(ctx context.Context, id uuid.UUID, userID uuid.UUID) (Collection, error)
	GetPublic(ctx context.Context, username string, slug string) (PublicCollection, error)
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

// Malformed usernames and slugs can't match any collection, so they are
// reported as not found without querying the database.
func (s *Service) GetPublic(ctx context.Context, username string, slug string) (PublicCollection, error) {
	username, err := validateUsername(username)
	if err != nil {
		return PublicCollection{}, ErrNotFound
	}
	slug, err = ValidateSlug(slug)
	if err != nil {
		return PublicCollection{}, ErrNotFound
	}
	return s.repo.GetPublic(ctx, username, slug)
}

func (s *Service) Clone(ctx context.Context, sourceID uuid.UUID, userID uuid.UUID) (Collection, error) {
	return s.repo.ClonePublic(ctx, sourceID, userID)
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
	slug, err := ValidateSlug(params.Slug)
	if err != nil {
		return Collection{}, err
	}

	return s.repo.Create(ctx, CreateParams{
		UserID:      params.UserID,
		Name:        name,
		Slug:        slug,
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
	slug, err := ValidateSlug(params.Slug)
	if err != nil {
		return Collection{}, err
	}

	return s.repo.Update(ctx, UpdateParams{
		ID:          params.ID,
		UserID:      params.UserID,
		Name:        name,
		Slug:        slug,
		Description: description,
		Public:      params.Public,
	})
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) (Collection, error) {
	return s.repo.Delete(ctx, id, userID)
}
