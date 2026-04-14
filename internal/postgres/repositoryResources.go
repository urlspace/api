package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/hreftools/api/internal/db"
	"github.com/hreftools/api/internal/resource"
)

type ResourceRepository struct {
	queries db.Querier
}

func NewResourceRepository(queries db.Querier) resource.Repository {
	return &ResourceRepository{queries: queries}
}

func toResource(r db.Resource) resource.Resource {
	return resource.Resource{
		ID:          r.ID,
		UserID:      r.UserID,
		Title:       r.Title,
		Description: r.Description,
		Url:         r.Url,
		Favourite:   r.Favourite,
		ReadLater:   r.ReadLater,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

func (r *ResourceRepository) List(ctx context.Context, userID uuid.UUID) ([]resource.Resource, error) {
	rows, err := r.queries.ListResources(ctx, userID)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, len(rows))
	for i, row := range rows {
		resources[i] = toResource(row)
	}
	return resources, nil
}

func (r *ResourceRepository) Get(ctx context.Context, id uuid.UUID, userID uuid.UUID) (resource.Resource, error) {
	row, err := r.queries.GetResource(ctx, db.GetResourceParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return resource.Resource{}, err
	}
	return toResource(row), nil
}

func (r *ResourceRepository) Create(ctx context.Context, params resource.CreateParams) (resource.Resource, error) {
	args := db.CreateResourceParams{
		UserID:      params.UserID,
		Title:       params.Title,
		Description: params.Description,
		Url:         params.Url,
		Favourite:   params.Favourite,
		ReadLater:   params.ReadLater,
	}
	row, err := r.queries.CreateResource(ctx, args)
	if err != nil {
		return resource.Resource{}, err
	}
	return toResource(row), nil
}

func (r *ResourceRepository) Update(ctx context.Context, params resource.UpdateParams) (resource.Resource, error) {
	args := db.UpdateResourceParams{
		ID:          params.ID,
		UserID:      params.UserID,
		Title:       params.Title,
		Description: params.Description,
		Url:         params.Url,
		Favourite:   params.Favourite,
		ReadLater:   params.ReadLater,
	}
	row, err := r.queries.UpdateResource(ctx, args)
	if err != nil {
		return resource.Resource{}, err
	}
	return toResource(row), nil
}

func (r *ResourceRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) (resource.Resource, error) {
	row, err := r.queries.DeleteResource(ctx, db.DeleteResourceParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return resource.Resource{}, err
	}
	return toResource(row), nil
}
