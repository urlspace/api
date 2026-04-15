package resource

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

type CreateParams struct {
	UserID      uuid.UUID
	Title       string
	Description string
	Url         string
	Favourite   bool
	ReadLater   bool
}

type UpdateParams struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Title       string
	Description string
	Url         string
	Favourite   bool
	ReadLater   bool
}

type Repository interface {
	List(ctx context.Context, userID uuid.UUID) ([]Resource, error)
	Get(ctx context.Context, id uuid.UUID, userID uuid.UUID) (Resource, error)
	Create(ctx context.Context, params CreateParams) (Resource, error)
	Update(ctx context.Context, params UpdateParams) (Resource, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) (Resource, error)
}

type Service struct {
	Repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]Resource, error) {
	return s.Repo.List(ctx, userID)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID, userID uuid.UUID) (Resource, error) {
	return s.Repo.Get(ctx, id, userID)
}

func (s *Service) Create(ctx context.Context, params CreateParams) (Resource, error) {
	return s.Repo.Create(ctx, params)
}

func (s *Service) Update(ctx context.Context, params UpdateParams) (Resource, error) {
	return s.Repo.Update(ctx, params)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) (Resource, error) {
	return s.Repo.Delete(ctx, id, userID)
}
