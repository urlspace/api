package uow

import (
	"context"

	"github.com/google/uuid"
	"github.com/urlspace/api/internal/collection"
	"github.com/urlspace/api/internal/link"
	"github.com/urlspace/api/internal/tag"
)

// Repositories groups all repositories available within a transaction. It lives here
// because the uow package coordinates across multiple domain repositories.
// Neither link nor tag imports this package.
type Repositories struct {
	Links       link.Repository
	Tags        tag.Repository
	Collections collection.Repository
}

// UnitOfWork runs fn inside a single database transaction. Every repository in
// the Repositories value passed to fn executes against that transaction.
type UnitOfWork interface {
	RunInTx(ctx context.Context, fn func(Repositories) error) error
}

type Service struct {
	// repos holds repositories for single-repo operations that don't need
	// transactional guarantees. More repositories can be added here as needed.
	repos Repositories
	// uow wraps multi-repo operations in a transaction. Currently used for
	// link + tag coordination only.
	uow UnitOfWork
}

func NewService(repos Repositories, uow UnitOfWork) *Service {
	return &Service{repos: repos, uow: uow}
}

// CollectionInfo is a lightweight summary of a collection, included in
// enriched link responses. Only ID and Title are needed for display.
type CollectionInfo struct {
	ID    uuid.UUID
	Title string
}

// EnrichedLink extends link.Link with tag and collection data.
// The link package stays independent of tags and collections, so this
// combined type lives here where all domains are coordinated.
type EnrichedLink struct {
	link.Link
	Tags       []string
	Collection *CollectionInfo
}

// collectionInfoFromLink builds a CollectionInfo from the link's
// JOIN-populated fields (used on Get/List paths).
func collectionInfoFromLink(l link.Link) *CollectionInfo {
	if l.CollectionID == nil {
		return nil
	}
	return &CollectionInfo{ID: *l.CollectionID, Title: l.CollectionTitle}
}

type CreateLinkParams struct {
	UserID       uuid.UUID
	Title        string
	Description  string
	URL          string
	CollectionID *uuid.UUID
	Tags         []string
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

	var result EnrichedLink

	err = s.uow.RunInTx(ctx, func(repos Repositories) error {
		// Validate collection ownership if provided.
		if params.CollectionID != nil {
			c, err := repos.Collections.Get(ctx, *params.CollectionID, params.UserID)
			if err != nil {
				return err
			}
			result.Collection = &CollectionInfo{ID: c.ID, Title: c.Title}
		}

		l, err := repos.Links.Create(ctx, link.CreateParams{
			UserID:       params.UserID,
			Title:        title,
			Description:  description,
			URL:          url,
			CollectionID: params.CollectionID,
		})
		if err != nil {
			return err
		}
		result.Link = l

		tagIDs := make([]uuid.UUID, 0, len(tagNames))
		for _, name := range tagNames {
			t, err := repos.Tags.UpsertByName(ctx, params.UserID, name)
			if err != nil {
				return err
			}
			tagIDs = append(tagIDs, t.ID)
		}

		if err := repos.Tags.ReplaceLinkTags(ctx, l.ID, tagIDs); err != nil {
			return err
		}

		tags, err := repos.Tags.GetTagsForLink(ctx, l.ID)
		if err != nil {
			return err
		}
		result.Tags = tags

		return nil
	})

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

	var result EnrichedLink

	err = s.uow.RunInTx(ctx, func(repos Repositories) error {
		// Validate collection ownership if provided.
		if params.CollectionID != nil {
			c, err := repos.Collections.Get(ctx, *params.CollectionID, params.UserID)
			if err != nil {
				return err
			}
			result.Collection = &CollectionInfo{ID: c.ID, Title: c.Title}
		}

		l, err := repos.Links.Update(ctx, link.UpdateParams{
			ID:           params.ID,
			UserID:       params.UserID,
			Title:        title,
			Description:  description,
			URL:          url,
			CollectionID: params.CollectionID,
		})
		if err != nil {
			return err
		}
		result.Link = l

		tagIDs := make([]uuid.UUID, 0, len(tagNames))
		for _, name := range tagNames {
			t, err := repos.Tags.UpsertByName(ctx, params.UserID, name)
			if err != nil {
				return err
			}
			tagIDs = append(tagIDs, t.ID)
		}

		if err := repos.Tags.ReplaceLinkTags(ctx, l.ID, tagIDs); err != nil {
			return err
		}

		tags, err := repos.Tags.GetTagsForLink(ctx, l.ID)
		if err != nil {
			return err
		}
		result.Tags = tags

		return nil
	})

	return result, err
}

func (s *Service) ListLinks(ctx context.Context, userID uuid.UUID) ([]EnrichedLink, error) {
	list, err := s.repos.Links.List(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(list) == 0 {
		return []EnrichedLink{}, nil
	}

	linkIDs := make([]uuid.UUID, len(list))
	for i, item := range list {
		linkIDs[i] = item.ID
	}

	tagsMap, err := s.repos.Tags.GetTagsForLinks(ctx, linkIDs)
	if err != nil {
		return nil, err
	}

	result := make([]EnrichedLink, len(list))
	for i, item := range list {
		tags := tagsMap[item.ID]
		if tags == nil {
			tags = []string{}
		}
		result[i] = EnrichedLink{
			Link:       item,
			Tags:       tags,
			Collection: collectionInfoFromLink(item),
		}
	}

	return result, nil
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
		Tags:       tags,
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
		Tags:       tags,
		Collection: col,
	}, nil
}
