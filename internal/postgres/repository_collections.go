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

// ClonePublic checks eligibility and copies the collection and links in one SQL
// statement, using one snapshot and an implicit transaction for atomicity.
// The query returns the new collection with cloned=true on success, or the
// source with cloned=false for an owned collection, without inserting anything.
// An unavailable source returns no row; a duplicate destination name or slug
// fails the statement. These outcomes map to ErrCloneOwnCollection, ErrNotFound and
// ErrConflict respectively.
func (r *CollectionRepository) ClonePublic(ctx context.Context, sourceID uuid.UUID, userID uuid.UUID) (collection.Collection, error) {
	row, err := r.queries.ClonePublicCollection(ctx, db.ClonePublicCollectionParams{
		SourceID: sourceID,
		UserID:   userID,
	})
	if err != nil {
		return collection.Collection{}, translateCollectionError(err)
	}

	if !row.Cloned {
		return collection.Collection{}, collection.ErrCloneOwnCollection
	}

	return collection.Collection{
		ID:          row.ID,
		UserID:      row.UserID,
		Name:        row.Name,
		Slug:        row.Slug,
		Description: row.Description,
		Public:      row.Public,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}

func translateCollectionError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return collection.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		if pgErr.ConstraintName == "collections_user_id_slug_key" {
			return collection.ErrSlugConflict
		}
		return collection.ErrConflict
	}
	return err
}

func toCollection(c db.Collection) collection.Collection {
	return collection.Collection{
		ID:          c.ID,
		UserID:      c.UserID,
		Name:        c.Name,
		Slug:        c.Slug,
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
				Slug:        row.Slug,
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

func (r *CollectionRepository) GetPublic(ctx context.Context, username string, slug string) (collection.PublicCollection, error) {
	rows, err := r.queries.GetPublicCollection(ctx, db.GetPublicCollectionParams{
		Username: username,
		Slug:     slug,
	})
	if err != nil {
		return collection.PublicCollection{}, translateCollectionError(err)
	}
	if len(rows) == 0 {
		return collection.PublicCollection{}, collection.ErrNotFound
	}

	first := rows[0]
	result := collection.PublicCollection{
		ID:          first.ID,
		Name:        first.Name,
		Slug:        first.Slug,
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
		Slug:        params.Slug,
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
		Slug:        params.Slug,
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
