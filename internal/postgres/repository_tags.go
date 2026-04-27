package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/urlspace/api/internal/db"
	"github.com/urlspace/api/internal/tag"
)

type TagRepository struct {
	queries db.Querier
}

func NewTagRepository(queries db.Querier) tag.Repository {
	return &TagRepository{queries: queries}
}

func translateTagError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return tag.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return tag.ErrConflict
	}
	return err
}

func toTag(t db.Tag) tag.Tag {
	return tag.Tag{
		ID:        t.ID,
		UserID:    t.UserID,
		Name:      t.Name,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

func (r *TagRepository) List(ctx context.Context, userID uuid.UUID) ([]tag.Tag, error) {
	rows, err := r.queries.ListTags(ctx, userID)
	if err != nil {
		return nil, translateTagError(err)
	}

	tags := make([]tag.Tag, len(rows))
	for i, row := range rows {
		tags[i] = toTag(row)
	}
	return tags, nil
}

func (r *TagRepository) Update(ctx context.Context, params tag.UpdateParams) (tag.Tag, error) {
	row, err := r.queries.UpdateTag(ctx, db.UpdateTagParams{
		ID:     params.ID,
		UserID: params.UserID,
		Name:   params.Name,
	})
	if err != nil {
		return tag.Tag{}, translateTagError(err)
	}
	return toTag(row), nil
}

func (r *TagRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) (tag.Tag, error) {
	row, err := r.queries.DeleteTag(ctx, db.DeleteTagParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return tag.Tag{}, translateTagError(err)
	}
	return toTag(row), nil
}

func (r *TagRepository) UpsertByName(ctx context.Context, userID uuid.UUID, name string) (tag.Tag, error) {
	// UpsertTag uses ON CONFLICT DO NOTHING, which means RETURNING returns no
	// rows when the tag already exists. In that case, fall back to GetTagByName.
	row, err := r.queries.UpsertTag(ctx, db.UpsertTagParams{
		UserID: userID,
		Name:   name,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			existing, err := r.queries.GetTagByName(ctx, db.GetTagByNameParams{
				UserID: userID,
				Name:   name,
			})
			if err != nil {
				return tag.Tag{}, translateTagError(err)
			}
			return toTag(existing), nil
		}
		return tag.Tag{}, translateTagError(err)
	}
	return toTag(row), nil
}

func (r *TagRepository) GetTagsForLink(ctx context.Context, linkID uuid.UUID) ([]string, error) {
	tags, err := r.queries.GetTagsForLink(ctx, linkID)
	if err != nil {
		return nil, translateTagError(err)
	}

	return tags, nil
}

func (r *TagRepository) GetTagsForLinks(ctx context.Context, linkIDs []uuid.UUID) (map[uuid.UUID][]string, error) {
	rows, err := r.queries.GetTagsForLinks(ctx, linkIDs)
	if err != nil {
		return nil, translateTagError(err)
	}

	result := make(map[uuid.UUID][]string, len(linkIDs))
	for _, row := range rows {
		result[row.LinkID] = append(result[row.LinkID], row.Name)
	}
	return result, nil
}

func (r *TagRepository) ReplaceLinkTags(ctx context.Context, linkID uuid.UUID, tagIDs []uuid.UUID) error {
	if err := r.queries.DeleteLinkTags(ctx, linkID); err != nil {
		return translateTagError(err)
	}

	for _, tagID := range tagIDs {
		if err := r.queries.CreateLinkTag(ctx, db.CreateLinkTagParams{
			LinkID: linkID,
			TagID:  tagID,
		}); err != nil {
			return translateTagError(err)
		}
	}

	return nil
}
