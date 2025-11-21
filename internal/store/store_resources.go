package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/zapi-sh/api/internal/db"
)

type ResourceStore interface {
	List(ctx context.Context) ([]db.Resource, error)
	Get(ctx context.Context, id uuid.UUID) (db.Resource, error)
	Create(ctx context.Context, title string, description string, url string, favourite bool, read_later bool) (db.Resource, error)
	Update(ctx context.Context, id uuid.UUID, title string, description string, url string, favourite bool, read_later bool) (db.Resource, error)
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

func (r *resourceStore) Create(ctx context.Context, title string, description string, url string, favourite bool, read_later bool) (db.Resource, error) {
	args := db.CreateResourceParams{
		Title:       title,
		Description: description,
		Url:         url,
		Favourite:   favourite,
		ReadLater:   read_later,
	}
	return r.queries.CreateResource(ctx, args)
}

func (r *resourceStore) Get(ctx context.Context, id uuid.UUID) (db.Resource, error) {
	return r.queries.GetResource(ctx, id)
}

func (r *resourceStore) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteResource(ctx, id)
}

func (r *resourceStore) Update(ctx context.Context, id uuid.UUID, title string, description string, url string, favourite bool, read_later bool) (db.Resource, error) {

	args := db.UpdateResourceParams{
		ID:          id,
		Title:       title,
		Description: description,
		Url:         url,
		Favourite:   favourite,
		ReadLater:   read_later,
	}

	return r.queries.UpdateResource(ctx, args)
}
