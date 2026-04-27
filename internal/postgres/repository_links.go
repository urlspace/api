package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/urlspace/api/internal/db"
	"github.com/urlspace/api/internal/link"
)

type LinkRepository struct {
	queries db.Querier
}

func NewLinkRepository(queries db.Querier) link.Repository {
	return &LinkRepository{queries: queries}
}

func translateLinkError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return link.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return link.ErrConflict
	}
	return err
}

func toCollectionID(n uuid.NullUUID) *uuid.UUID {
	if n.Valid {
		return &n.UUID
	}
	return nil
}

func toNullUUID(id *uuid.UUID) uuid.NullUUID {
	if id != nil {
		return uuid.NullUUID{UUID: *id, Valid: true}
	}
	return uuid.NullUUID{}
}

// toLink maps a db.Link to a domain Link. Used by Create, Update,
// and Delete which return plain table columns via RETURNING *. Get and List
// use a custom mapping because their LEFT JOIN returns additional columns
// (CollectionTitle) not present in db.Link.
func toLink(l db.Link) link.Link {
	return link.Link{
		ID:           l.ID,
		UserID:       l.UserID,
		Title:        l.Title,
		Description:  l.Description,
		URL:          l.Url,
		CollectionID: toCollectionID(l.CollectionID),
		CreatedAt:    l.CreatedAt,
		UpdatedAt:    l.UpdatedAt,
	}
}

func (r *LinkRepository) List(ctx context.Context, userID uuid.UUID) ([]link.Link, error) {
	rows, err := r.queries.ListLinks(ctx, userID)
	if err != nil {
		return nil, translateLinkError(err)
	}

	links := make([]link.Link, len(rows))
	for i, row := range rows {
		links[i] = link.Link{
			ID:              row.ID,
			UserID:          row.UserID,
			Title:           row.Title,
			Description:     row.Description,
			URL:             row.Url,
			CollectionID:    toCollectionID(row.CollectionID),
			CollectionTitle: row.CollectionTitle.String,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
		}
	}
	return links, nil
}

func (r *LinkRepository) Get(ctx context.Context, id uuid.UUID, userID uuid.UUID) (link.Link, error) {
	row, err := r.queries.GetLink(ctx, db.GetLinkParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return link.Link{}, translateLinkError(err)
	}
	return link.Link{
		ID:              row.ID,
		UserID:          row.UserID,
		Title:           row.Title,
		Description:     row.Description,
		URL:             row.Url,
		CollectionID:    toCollectionID(row.CollectionID),
		CollectionTitle: row.CollectionTitle.String,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}, nil
}

func (r *LinkRepository) Create(ctx context.Context, params link.CreateParams) (link.Link, error) {
	args := db.CreateLinkParams{
		UserID:       params.UserID,
		Title:        params.Title,
		Description:  params.Description,
		Url:          params.URL,
		CollectionID: toNullUUID(params.CollectionID),
	}
	row, err := r.queries.CreateLink(ctx, args)
	if err != nil {
		return link.Link{}, translateLinkError(err)
	}
	return toLink(row), nil
}

func (r *LinkRepository) Update(ctx context.Context, params link.UpdateParams) (link.Link, error) {
	args := db.UpdateLinkParams{
		ID:           params.ID,
		UserID:       params.UserID,
		Title:        params.Title,
		Description:  params.Description,
		Url:          params.URL,
		CollectionID: toNullUUID(params.CollectionID),
	}
	row, err := r.queries.UpdateLink(ctx, args)
	if err != nil {
		return link.Link{}, translateLinkError(err)
	}
	return toLink(row), nil
}

func (r *LinkRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) (link.Link, error) {
	row, err := r.queries.DeleteLink(ctx, db.DeleteLinkParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return link.Link{}, translateLinkError(err)
	}
	return toLink(row), nil
}
