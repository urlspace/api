package tag

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	// validation errors
	ErrValidationNameLength     = errors.New("tag name must be between 2 and 50 characters")
	ErrValidationNameCharacters = errors.New("tag name must contain only lowercase letters, digits, and hyphens")
	ErrValidationNameHyphens    = errors.New("tag name must not start, end with, or contain consecutive hyphens")
	ErrValidationTooManyTags    = errors.New("a link can have at most 10 tags")

	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

type Repository interface {
	List(ctx context.Context, userID uuid.UUID) ([]Tag, error)
	Update(ctx context.Context, params UpdateParams) (Tag, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) (Tag, error)
	UpsertByName(ctx context.Context, userID uuid.UUID, name string) (Tag, error)
	GetTagsForLink(ctx context.Context, linkID uuid.UUID) ([]string, error)
	GetTagsForLinks(ctx context.Context, linkIDs []uuid.UUID) (map[uuid.UUID][]string, error)
	ReplaceLinkTags(ctx context.Context, linkID uuid.UUID, tagIDs []uuid.UUID) error
}

type UpdateParams struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Name   string
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]Tag, error) {
	return s.repo.List(ctx, userID)
}

func (s *Service) Update(ctx context.Context, params UpdateParams) (Tag, error) {
	name, err := validateName(params.Name)
	if err != nil {
		return Tag{}, err
	}

	repoParams := UpdateParams{
		ID:     params.ID,
		UserID: params.UserID,
		Name:   name,
	}
	return s.repo.Update(ctx, repoParams)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) (Tag, error) {
	return s.repo.Delete(ctx, id, userID)
}
