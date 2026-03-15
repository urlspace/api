package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/hreftools/api/internal/db"
)

type ResourceCreateParams struct {
	UserID      uuid.UUID
	Title       string
	Description string
	Url         string
	Favourite   bool
	ReadLater   bool
}

type ResourceUpdateParams struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Title       string
	Description string
	Url         string
	Favourite   bool
	ReadLater   bool
}

type ResourceStore interface {
	List(ctx context.Context, userID uuid.UUID) ([]db.Resource, error)
	Get(ctx context.Context, id uuid.UUID, userID uuid.UUID) (db.Resource, error)
	Create(ctx context.Context, params ResourceCreateParams) (db.Resource, error)
	Update(ctx context.Context, params ResourceUpdateParams) (db.Resource, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) (db.Resource, error)
}

type resourceStore struct {
	queries db.Querier
}

func NewResourceStore(queries db.Querier) ResourceStore {
	return &resourceStore{
		queries: queries,
	}
}

func (r *resourceStore) List(ctx context.Context, userID uuid.UUID) ([]db.Resource, error) {
	return r.queries.ListResources(ctx, userID)
}

func (r *resourceStore) Create(ctx context.Context, params ResourceCreateParams) (db.Resource, error) {
	args := db.CreateResourceParams{
		UserID:      params.UserID,
		Title:       params.Title,
		Description: params.Description,
		Url:         params.Url,
		Favourite:   params.Favourite,
		ReadLater:   params.ReadLater,
	}
	return r.queries.CreateResource(ctx, args)
}

func (r *resourceStore) Get(ctx context.Context, id uuid.UUID, userID uuid.UUID) (db.Resource, error) {
	return r.queries.GetResource(ctx, db.GetResourceParams{
		ID:     id,
		UserID: userID,
	})
}

func (r *resourceStore) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) (db.Resource, error) {
	return r.queries.DeleteResource(ctx, db.DeleteResourceParams{
		ID:     id,
		UserID: userID,
	})
}

func (r *resourceStore) Update(ctx context.Context, params ResourceUpdateParams) (db.Resource, error) {
	args := db.UpdateResourceParams{
		ID:          params.ID,
		UserID:      params.UserID,
		Title:       params.Title,
		Description: params.Description,
		Url:         params.Url,
		Favourite:   params.Favourite,
		ReadLater:   params.ReadLater,
	}

	return r.queries.UpdateResource(ctx, args)
}
