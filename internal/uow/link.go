package uow

import (
	"context"
	"math"
	"strings"

	"github.com/urlspace/api/internal/config"
	"github.com/urlspace/api/internal/link"
	"github.com/urlspace/api/internal/tag"
	"go.opentelemetry.io/otel/codes"
	"uuid"
)

// CollectionInfo is a lightweight summary of a collection, included in
// enriched link responses. Only ID and Name are needed for display.
type CollectionInfo struct {
	ID   uuid.UUID
	Name string
}

// TagInfo is a lightweight summary of a tag, included in enriched link
// responses. Only ID and Name are needed for display.
type TagInfo struct {
	ID   uuid.UUID
	Name string
}

// EnrichedLink extends link.Link with tag and collection data.
// The link package stays independent of tags and collections, so this
// combined type lives here where all domains are coordinated.
type EnrichedLink struct {
	link.Link
	Tags       []TagInfo
	Collection *CollectionInfo
}

// collectionInfoFromLink builds a CollectionInfo from the link's
// JOIN-populated fields (used on Get/List paths).
func collectionInfoFromLink(l link.Link) *CollectionInfo {
	if l.CollectionID == nil {
		return nil
	}
	return &CollectionInfo{ID: *l.CollectionID, Name: l.CollectionName}
}

// tagInfosFromTags projects full tag.Tag values to the lightweight TagInfo
// shape used in enriched link responses.
func tagInfosFromTags(tags []tag.Tag) []TagInfo {
	result := make([]TagInfo, len(tags))
	for i, t := range tags {
		result[i] = TagInfo{ID: t.ID, Name: t.Name}
	}
	return result
}

// createTags creates the missing tags from names and enforces the free tags
// limit. It returns how many tags were created.
func createTags(ctx context.Context, repos Repositories, userID uuid.UUID, names []string, unlimited bool) (int, error) {
	if len(names) == 0 {
		return 0, nil
	}

	var countBefore int
	if !unlimited {
		var err error
		countBefore, err = repos.Tags.Count(ctx, userID)
		if err != nil {
			return 0, err
		}
	}

	created, err := repos.Tags.CreateMany(ctx, userID, names)
	if err != nil {
		return 0, err
	}

	if !unlimited && created > 0 && countBefore+created > config.FreeTagsLimit {
		return 0, tag.ErrLimitReached
	}

	return created, nil
}

// setLinkTags creates the missing tags from names and replaces the link's
// tags with them. It returns the tags now attached to the link.
func setLinkTags(ctx context.Context, repos Repositories, userID, linkID uuid.UUID, names []string, unlimited bool) ([]tag.Tag, error) {
	if _, err := createTags(ctx, repos, userID, names, unlimited); err != nil {
		return nil, err
	}

	tags := []tag.Tag{}
	if len(names) > 0 {
		var err error
		tags, err = repos.Tags.GetByNames(ctx, userID, names)
		if err != nil {
			return nil, err
		}
	}

	tagIDs := make([]uuid.UUID, len(tags))
	for i, t := range tags {
		tagIDs[i] = t.ID
	}

	if err := repos.Tags.ReplaceLinkTags(ctx, linkID, tagIDs); err != nil {
		return nil, err
	}

	return tags, nil
}

// normalizeLikePattern trims whitespace and escapes the LIKE wildcards
// %, _, and \ so user input is treated as literal text. Without escaping,
// searching for "100%" matches any title containing "100", and "foo_bar"
// matches "foo<X>bar". Postgres uses backslash as the default LIKE escape,
// so no ESCAPE clause is needed.
func normalizeLikePattern(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

type CreateLinkParams struct {
	UserID       uuid.UUID
	Title        string
	Description  string
	URL          string
	CollectionID *uuid.UUID
	Tags         []string
	Favourite    bool
	ForLater     bool
	Unlimited    bool
}

func (s *Service) CreateLink(ctx context.Context, params CreateLinkParams) (EnrichedLink, error) {
	title, err := link.ValidateTitle(params.Title)
	if err != nil {
		return EnrichedLink{}, err
	}
	description, err := link.ValidateDescription(params.Description)
	if err != nil {
		return EnrichedLink{}, err
	}
	url, err := link.ValidateURL(params.URL)
	if err != nil {
		return EnrichedLink{}, err
	}
	tagNames, err := tag.ValidateTagNames(params.Tags)
	if err != nil {
		return EnrichedLink{}, err
	}

	ctx, span := tracer.Start(ctx, "uow.CreateLink")
	defer span.End()

	var result EnrichedLink

	err = s.uow.RunInTx(ctx, func(repos Repositories) error {
		if !params.Unlimited {
			count, err := repos.Links.Count(ctx, link.ListFilter{UserID: params.UserID})
			if err != nil {
				return err
			}
			if count >= config.FreeLinksLimit {
				return link.ErrLimitReached
			}
		}

		// Validate collection ownership if provided.
		if params.CollectionID != nil {
			c, err := repos.Collections.Get(ctx, *params.CollectionID, params.UserID)
			if err != nil {
				return err
			}
			result.Collection = &CollectionInfo{ID: c.ID, Name: c.Name}
		}

		l, err := repos.Links.Create(ctx, link.CreateParams{
			UserID:       params.UserID,
			Title:        title,
			Description:  description,
			URL:          url,
			CollectionID: params.CollectionID,
			Favourite:    params.Favourite,
			ForLater:     params.ForLater,
		})
		if err != nil {
			return err
		}
		result.Link = l

		tags, err := setLinkTags(ctx, repos, params.UserID, l.ID, tagNames, params.Unlimited)
		if err != nil {
			return err
		}
		result.Tags = tagInfosFromTags(tags)

		return nil
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return result, err
}

type UpdateLinkParams struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	Title        string
	Description  string
	URL          string
	CollectionID *uuid.UUID
	Tags         []string
	Favourite    bool
	ForLater     bool
	Unlimited    bool
}

func (s *Service) UpdateLink(ctx context.Context, params UpdateLinkParams) (EnrichedLink, error) {
	title, err := link.ValidateTitle(params.Title)
	if err != nil {
		return EnrichedLink{}, err
	}
	description, err := link.ValidateDescription(params.Description)
	if err != nil {
		return EnrichedLink{}, err
	}
	url, err := link.ValidateURL(params.URL)
	if err != nil {
		return EnrichedLink{}, err
	}
	tagNames, err := tag.ValidateTagNames(params.Tags)
	if err != nil {
		return EnrichedLink{}, err
	}

	ctx, span := tracer.Start(ctx, "uow.UpdateLink")
	defer span.End()

	var result EnrichedLink

	err = s.uow.RunInTx(ctx, func(repos Repositories) error {
		// Validate collection ownership if provided.
		if params.CollectionID != nil {
			c, err := repos.Collections.Get(ctx, *params.CollectionID, params.UserID)
			if err != nil {
				return err
			}
			result.Collection = &CollectionInfo{ID: c.ID, Name: c.Name}
		}

		l, err := repos.Links.Update(ctx, link.UpdateParams{
			ID:           params.ID,
			UserID:       params.UserID,
			Title:        title,
			Description:  description,
			URL:          url,
			CollectionID: params.CollectionID,
			Favourite:    params.Favourite,
			ForLater:     params.ForLater,
		})
		if err != nil {
			return err
		}
		result.Link = l

		tags, err := setLinkTags(ctx, repos, params.UserID, l.ID, tagNames, params.Unlimited)
		if err != nil {
			return err
		}
		result.Tags = tagInfosFromTags(tags)

		return nil
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return result, err
}

type ListLinksParams struct {
	UserID       uuid.UUID
	Page         int
	PageSize     int
	CollectionID *uuid.UUID
	TagIDs       []uuid.UUID
	Query        string
	Favourite    *bool
	ForLater     *bool
}

type ListLinksResult struct {
	Links      []EnrichedLink
	TotalCount int
}

func (s *Service) ListLinks(ctx context.Context, params ListLinksParams) (ListLinksResult, error) {
	ctx, span := tracer.Start(ctx, "uow.ListLinks")
	defer span.End()

	// Decision (future me): dedupe TagIDs. The SQL compares
	// COUNT(DISTINCT tag_id) to cardinality($4), so duplicates in the input
	// would inflate the required count and exclude matching links.
	var tagIDs []uuid.UUID
	if len(params.TagIDs) > 0 {
		seen := make(map[uuid.UUID]struct{}, len(params.TagIDs))
		tagIDs = make([]uuid.UUID, 0, len(params.TagIDs))
		for _, id := range params.TagIDs {
			if _, ok := seen[id]; !ok {
				seen[id] = struct{}{}
				tagIDs = append(tagIDs, id)
			}
		}
	}

	filter := link.ListFilter{
		UserID:       params.UserID,
		CollectionID: params.CollectionID,
		TagIDs:       tagIDs,
		Query:        normalizeLikePattern(params.Query),
		Favourite:    params.Favourite,
		ForLater:     params.ForLater,
	}

	totalCount, err := s.repos.Links.Count(ctx, filter)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return ListLinksResult{}, err
	}

	// Decision (future me): three cases collapse into one short-circuit:
	// (1) empty table, (2) well-formed request for a page beyond totalPages,
	// (3) pathologically large Page that would overflow the int multiplication
	// and emit garbage to the int32 cast at the repo layer. Skip the data
	// query — no rows to fetch. The overflow check uses MaxInt/PageSize as
	// the largest safe (Page-1); above that the multiplication wraps. The
	// offset value computed in the overflow case is discarded.
	offset := (params.Page - 1) * params.PageSize
	overflow := params.Page > 1 && params.Page-1 > math.MaxInt/params.PageSize
	if overflow || offset >= totalCount {
		return ListLinksResult{Links: []EnrichedLink{}, TotalCount: totalCount}, nil
	}

	list, err := s.repos.Links.List(ctx, filter, params.PageSize, offset)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return ListLinksResult{}, err
	}

	// Decision (future me): race-safety net, not redundant with the offset
	// check above. If a link is deleted between Count and List, the requested
	// page can come back empty even though offset was in range. Skip the
	// tag-enrichment join — calling GetTagsForLinks with an empty slice is
	// wasted work. Super edge casy, but not impossible.
	if len(list) == 0 {
		return ListLinksResult{Links: []EnrichedLink{}, TotalCount: totalCount}, nil
	}

	linkIDs := make([]uuid.UUID, len(list))
	for i, item := range list {
		linkIDs[i] = item.ID
	}

	tagsMap, err := s.repos.Tags.GetTagsForLinks(ctx, linkIDs)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return ListLinksResult{}, err
	}

	links := make([]EnrichedLink, len(list))
	for i, item := range list {
		links[i] = EnrichedLink{
			Link:       item,
			Tags:       tagInfosFromTags(tagsMap[item.ID]),
			Collection: collectionInfoFromLink(item),
		}
	}

	return ListLinksResult{Links: links, TotalCount: totalCount}, nil
}

func (s *Service) GetLink(ctx context.Context, id uuid.UUID, userID uuid.UUID) (EnrichedLink, error) {
	l, err := s.repos.Links.Get(ctx, id, userID)
	if err != nil {
		return EnrichedLink{}, err
	}

	tags, err := s.repos.Tags.GetTagsForLink(ctx, id)
	if err != nil {
		return EnrichedLink{}, err
	}

	return EnrichedLink{
		Link:       l,
		Tags:       tagInfosFromTags(tags),
		Collection: collectionInfoFromLink(l),
	}, nil
}

func (s *Service) DeleteLink(ctx context.Context, id uuid.UUID, userID uuid.UUID) (EnrichedLink, error) {
	tags, err := s.repos.Tags.GetTagsForLink(ctx, id)
	if err != nil {
		return EnrichedLink{}, err
	}

	// Look up collection info before deleting (DELETE can't JOIN).
	l, err := s.repos.Links.Get(ctx, id, userID)
	if err != nil {
		return EnrichedLink{}, err
	}
	col := collectionInfoFromLink(l)

	deleted, err := s.repos.Links.Delete(ctx, id, userID)
	if err != nil {
		return EnrichedLink{}, err
	}

	return EnrichedLink{
		Link:       deleted,
		Tags:       tagInfosFromTags(tags),
		Collection: col,
	}, nil
}
