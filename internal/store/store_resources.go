package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/hreftools/api/internal/db"
)

type ResourceCreateParams struct {
	Title       string
	Description string
	Url         string
	Favourite   bool
	ReadLater   bool
}

type ResourceUpdateParams struct {
	ID          uuid.UUID
	Title       string
	Description string
	Url         string
	Favourite   bool
	ReadLater   bool
}

type ResourceStore interface {
	List(ctx context.Context) ([]db.Resource, error)
	Get(ctx context.Context, id uuid.UUID) (db.Resource, error)
	Create(ctx context.Context, params ResourceCreateParams) (db.Resource, error)
	Update(ctx context.Context, params ResourceUpdateParams) (db.Resource, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type resourceStore struct {
	queries db.Querier
}

func NewResourceStore(queries db.Querier) ResourceStore {
	return &resourceStore{
		queries: queries,
	}
}

func (r *resourceStore) List(ctx context.Context) ([]db.Resource, error) {
	return r.queries.ListResources(ctx)
}

func (r *resourceStore) Create(ctx context.Context, params ResourceCreateParams) (db.Resource, error) {
	args := db.CreateResourceParams{
		Title:       params.Title,
		Description: params.Description,
		Url:         params.Url,
		Favourite:   params.Favourite,
		ReadLater:   params.ReadLater,
	}
	return r.queries.CreateResource(ctx, args)
}

func (r *resourceStore) Get(ctx context.Context, id uuid.UUID) (db.Resource, error) {
	return r.queries.GetResource(ctx, id)
}

func (r *resourceStore) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteResource(ctx, id)
}

func (r *resourceStore) Update(ctx context.Context, params ResourceUpdateParams) (db.Resource, error) {

	args := db.UpdateResourceParams{
		ID:          params.ID,
		Title:       params.Title,
		Description: params.Description,
		Url:         params.Url,
		Favourite:   params.Favourite,
		ReadLater:   params.ReadLater,
	}

	return r.queries.UpdateResource(ctx, args)
}
