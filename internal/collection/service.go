package collection

import (
	"context"
	"crypto/subtle"
	"errors"

	"github.com/urlspace/api/internal/config"
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

	// Password validation errors.
	ErrValidationPasswordLength            = errors.New("password must be between 6 and 64 characters")
	ErrValidationPasswordInvalidCharacters = errors.New("password must not contain control characters")

	ErrNotFound           = errors.New("not found")
	ErrConflict           = errors.New("conflict")
	ErrSlugConflict       = errors.New("slug is already taken")
	ErrCloneOwnCollection = errors.New("you cannot clone your own collection")
	ErrLimitReached       = errors.New("collection limit reached")
	ErrLinkLimitReached   = errors.New("link limit reached")
	ErrPasswordRequired   = errors.New("password required")
	ErrInvalidPassword    = errors.New("invalid password")
)

type CloneCounts struct {
	Collections int
	Links       int
	SourceLinks int
}

type CreateParams struct {
	UserID      uuid.UUID
	Name        string
	Slug        string
	Description string
	Public      bool
	Password    *string
}

type UpdateParams struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Name        string
	Slug        string
	Description string
	Public      bool
	Password    *string
}

type Repository interface {
	ClonePublic(ctx context.Context, sourceID uuid.UUID, userID uuid.UUID, name string) (Collection, error)
	List(ctx context.Context, userID uuid.UUID) ([]CollectionWithLinkCount, error)
	Count(ctx context.Context, userID uuid.UUID) (int, error)
	GetCloneCounts(ctx context.Context, sourceID uuid.UUID, userID uuid.UUID) (CloneCounts, error)
	GetCloneSourcePassword(ctx context.Context, sourceID uuid.UUID) (*string, error)
	Get(ctx context.Context, id uuid.UUID, userID uuid.UUID) (Collection, error)
	GetPublic(ctx context.Context, username string, slug string) (PublicCollection, error)
	Create(ctx context.Context, params CreateParams) (Collection, error)
	Update(ctx context.Context, params UpdateParams) (Collection, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID, cascade bool) (Collection, error)
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

// checkPassword compares in constant time, so response timing doesn't reveal
// how much of a guess was right.
func checkPassword(stored *string, given *string) error {
	if stored == nil {
		return nil
	}
	if given == nil {
		return ErrPasswordRequired
	}
	if subtle.ConstantTimeCompare([]byte(*stored), []byte(*given)) != 1 {
		return ErrInvalidPassword
	}
	return nil
}

// Malformed usernames and slugs can't match any collection, so they are
// reported as not found without querying the database. A password protected
// collection is only returned with the right password.
func (s *Service) GetPublic(ctx context.Context, username string, slug string, password *string) (PublicCollection, error) {
	username, err := validateUsername(username)
	if err != nil {
		return PublicCollection{}, ErrNotFound
	}
	slug, err = ValidateSlug(slug)
	if err != nil {
		return PublicCollection{}, ErrNotFound
	}
	result, err := s.repo.GetPublic(ctx, username, slug)
	if err != nil {
		return PublicCollection{}, err
	}
	if err := checkPassword(result.Password, password); err != nil {
		return PublicCollection{}, err
	}
	return result, nil
}

func (s *Service) Clone(ctx context.Context, sourceID uuid.UUID, userID uuid.UUID, name string, password *string, unlimited bool) (Collection, error) {
	name, err := ValidateName(name)
	if err != nil {
		return Collection{}, err
	}

	sourcePassword, err := s.repo.GetCloneSourcePassword(ctx, sourceID)
	if err != nil {
		return Collection{}, err
	}
	if err := checkPassword(sourcePassword, password); err != nil {
		return Collection{}, err
	}

	if !unlimited {
		counts, err := s.repo.GetCloneCounts(ctx, sourceID, userID)
		if err != nil {
			return Collection{}, err
		}
		if counts.Collections >= config.FreeCollectionsLimit {
			return Collection{}, ErrLimitReached
		}
		if counts.Links+counts.SourceLinks > config.FreeLinksLimit {
			return Collection{}, ErrLinkLimitReached
		}
	}
	return s.repo.ClonePublic(ctx, sourceID, userID, name)
}

func (s *Service) Create(ctx context.Context, params CreateParams, unlimited bool) (Collection, error) {
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
	password, err := ValidatePassword(params.Password, params.Public)
	if err != nil {
		return Collection{}, err
	}

	if !unlimited {
		count, err := s.repo.Count(ctx, params.UserID)
		if err != nil {
			return Collection{}, err
		}
		if count >= config.FreeCollectionsLimit {
			return Collection{}, ErrLimitReached
		}
	}

	return s.repo.Create(ctx, CreateParams{
		UserID:      params.UserID,
		Name:        name,
		Slug:        slug,
		Description: description,
		Public:      params.Public,
		Password:    password,
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
	password, err := ValidatePassword(params.Password, params.Public)
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
		Password:    password,
	})
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID, cascade bool) (Collection, error) {
	return s.repo.Delete(ctx, id, userID, cascade)
}
