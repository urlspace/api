package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/urlspace/api/internal/collection"
	"github.com/urlspace/api/internal/db"
	"uuid"
)

type CollectionRepository struct {
	queries db.Querier
}

func NewCollectionRepository(queries db.Querier) collection.Repository {
	return &CollectionRepository{queries: queries}
}

func translateCollectionError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return collection.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return collection.ErrConflict
	}
	return err
}

func toCollection(c db.Collection) collection.Collection {
	return collection.Collection{
		ID:          c.ID,
		UserID:      c.UserID,
		Name:        c.Name,
		Description: c.Description,
		Public:      c.Public,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

func (r *CollectionRepository) List(ctx context.Context, userID uuid.UUID) ([]collection.CollectionWithLinkCount, error) {
	rows, err := r.queries.ListCollections(ctx, userID)
	if err != nil {
		return nil, translateCollectionError(err)
	}

	collections := make([]collection.CollectionWithLinkCount, len(rows))
	for i, row := range rows {
		collections[i] = collection.CollectionWithLinkCount{
			Collection: collection.Collection{
				ID:          row.ID,
				UserID:      row.UserID,
				Name:        row.Name,
				Description: row.Description,
				Public:      row.Public,
				CreatedAt:   row.CreatedAt,
				UpdatedAt:   row.UpdatedAt,
			},
			LinkCount: int(row.LinkCount),
		}
	}
	return collections, nil
}

func (r *CollectionRepository) Get(ctx context.Context, id uuid.UUID, userID uuid.UUID) (collection.Collection, error) {
	row, err := r.queries.GetCollection(ctx, db.GetCollectionParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return collection.Collection{}, translateCollectionError(err)
	}
	return toCollection(row), nil
}

func (r *CollectionRepository) GetPublic(ctx context.Context, id uuid.UUID) (collection.PublicCollection, error) {
	rows, err := r.queries.GetPublicCollection(ctx, id)
	if err != nil {
		return collection.PublicCollection{}, translateCollectionError(err)
	}
	if len(rows) == 0 {
		return collection.PublicCollection{}, collection.ErrNotFound
	}

	first := rows[0]
	result := collection.PublicCollection{
		Name:        first.Name,
		Description: first.Description,
		CreatedAt:   first.CreatedAt,
		UpdatedAt:   first.UpdatedAt,
		Author: collection.PublicAuthor{
			DisplayName: first.DisplayName,
			Username:    first.Username,
		},
		Links: make([]collection.PublicLink, 0, len(rows)),
	}
	for _, row := range rows {
		if row.LinkID == nil {
			continue
		}
		result.Links = append(result.Links, collection.PublicLink{
			ID:          *row.LinkID,
			Title:       *row.LinkTitle,
			Description: *row.LinkDescription,
			CreatedAt:   *row.LinkCreatedAt,
			URL:         *row.LinkUrl,
		})
	}
	return result, nil
}

func (r *CollectionRepository) Create(ctx context.Context, params collection.CreateParams) (collection.Collection, error) {
	row, err := r.queries.CreateCollection(ctx, db.CreateCollectionParams{
		UserID:      params.UserID,
		Name:        params.Name,
		Description: params.Description,
		Public:      params.Public,
	})
	if err != nil {
		return collection.Collection{}, translateCollectionError(err)
	}
	return toCollection(row), nil
}

func (r *CollectionRepository) Update(ctx context.Context, params collection.UpdateParams) (collection.Collection, error) {
	row, err := r.queries.UpdateCollection(ctx, db.UpdateCollectionParams{
		ID:          params.ID,
		UserID:      params.UserID,
		Name:        params.Name,
		Description: params.Description,
		Public:      params.Public,
	})
	if err != nil {
		return collection.Collection{}, translateCollectionError(err)
	}
	return toCollection(row), nil
}

func (r *CollectionRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) (collection.Collection, error) {
	row, err := r.queries.DeleteCollection(ctx, db.DeleteCollectionParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return collection.Collection{}, translateCollectionError(err)
	}
	return toCollection(row), nil
}
