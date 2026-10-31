package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/urlspace/api/internal/db"
	"github.com/urlspace/api/internal/tag"
	"uuid"
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

func (r *TagRepository) Count(ctx context.Context, userID uuid.UUID) (int, error) {
	count, err := r.queries.CountTags(ctx, userID)
	if err != nil {
		return 0, translateTagError(err)
	}
	return int(count), nil
}

func (r *TagRepository) List(ctx context.Context, userID uuid.UUID) ([]tag.TagWithLinkCount, error) {
	rows, err := r.queries.ListTags(ctx, userID)
	if err != nil {
		return nil, translateTagError(err)
	}

	tags := make([]tag.TagWithLinkCount, len(rows))
	for i, row := range rows {
		tags[i] = tag.TagWithLinkCount{
			Tag: tag.Tag{
				ID:        row.ID,
				UserID:    row.UserID,
				Name:      row.Name,
				CreatedAt: row.CreatedAt,
				UpdatedAt: row.UpdatedAt,
			},
			LinkCount: int(row.LinkCount),
		}
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

func (r *TagRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID, cascade bool) (tag.Tag, error) {
	var row db.Tag
	var err error
	if cascade {
		row, err = r.queries.DeleteTagCascade(ctx, db.DeleteTagCascadeParams{
			ID:     id,
			UserID: userID,
		})
	} else {
		row, err = r.queries.DeleteTag(ctx, db.DeleteTagParams{
			ID:     id,
			UserID: userID,
		})
	}
	if err != nil {
		return tag.Tag{}, translateTagError(err)
	}
	return toTag(row), nil
}

// CreateMany inserts the tags that don't exist yet and skips the rest. It
// returns how many tags were actually created.
func (r *TagRepository) CreateMany(ctx context.Context, userID uuid.UUID, names []string) (int, error) {
	created, err := r.queries.CreateTags(ctx, db.CreateTagsParams{
		UserID: userID,
		Names:  names,
	})
	if err != nil {
		return 0, translateTagError(err)
	}
	return int(created), nil
}

func (r *TagRepository) GetByNames(ctx context.Context, userID uuid.UUID, names []string) ([]tag.Tag, error) {
	rows, err := r.queries.GetTagsByNames(ctx, db.GetTagsByNamesParams{
		UserID: userID,
		Names:  names,
	})
	if err != nil {
		return nil, translateTagError(err)
	}

	tags := make([]tag.Tag, len(rows))
	for i, row := range rows {
		tags[i] = toTag(row)
	}
	return tags, nil
}

func (r *TagRepository) GetTagsForLink(ctx context.Context, linkID uuid.UUID) ([]tag.Tag, error) {
	rows, err := r.queries.GetTagsForLink(ctx, linkID)
	if err != nil {
		return nil, translateTagError(err)
	}

	tags := make([]tag.Tag, len(rows))
	for i, row := range rows {
		tags[i] = toTag(row)
	}
	return tags, nil
}

func (r *TagRepository) GetTagsForLinks(ctx context.Context, linkIDs []uuid.UUID) (map[uuid.UUID][]tag.Tag, error) {
	rows, err := r.queries.GetTagsForLinks(ctx, linkIDs)
	if err != nil {
		return nil, translateTagError(err)
	}

	result := make(map[uuid.UUID][]tag.Tag, len(linkIDs))
	for _, row := range rows {
		result[row.LinkID] = append(result[row.LinkID], tag.Tag{
			ID:        row.ID,
			UserID:    row.UserID,
			Name:      row.Name,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}
	return result, nil
}

func (r *TagRepository) ReplaceLinkTags(ctx context.Context, linkID uuid.UUID, tagIDs []uuid.UUID) error {
	if err := r.queries.DeleteLinkTags(ctx, linkID); err != nil {
		return translateTagError(err)
	}

	if len(tagIDs) == 0 {
		return nil
	}

	if err := r.queries.CreateLinkTags(ctx, db.CreateLinkTagsParams{
		LinkID: linkID,
		TagIds: tagIDs,
	}); err != nil {
		return translateTagError(err)
	}

	return nil
}
