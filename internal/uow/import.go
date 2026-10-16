package uow

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/urlspace/api/internal/collection"
	"github.com/urlspace/api/internal/config"
	"github.com/urlspace/api/internal/link"
	"github.com/urlspace/api/internal/tag"
	"go.opentelemetry.io/otel/codes"
	"uuid"
)

var (
	ErrImportEmpty             = errors.New("nothing to import")
	ErrImportUnknownCollection = errors.New("collection is not listed in collections")
)

type ImportCollection struct {
	Name        string
	Slug        string
	Description string
}

type ImportLink struct {
	Title       string
	Description string
	URL         string
	Tags        []string
	Favourite   bool
	ForLater    bool
	CreatedAt   *time.Time
	Collection  *string
}

type ImportParams struct {
	UserID      uuid.UUID
	Collections []ImportCollection
	Links       []ImportLink
	Unlimited   bool
}

type ImportResult struct {
	Links       int
	Collections int
	Tags        int
}

type importCollection struct {
	key         string
	name        string
	slug        string
	description string
}

type importLink struct {
	title         string
	description   string
	url           string
	tags          []string
	favourite     bool
	forLater      bool
	createdAt     time.Time
	collectionKey string
}

func (s *Service) Import(ctx context.Context, params ImportParams) (ImportResult, error) {
	// no links? early return
	if len(params.Links) == 0 {
		return ImportResult{}, ErrImportEmpty
	}

	// validate collections data and collect unique collections
	collections := make([]importCollection, 0, len(params.Collections))
	collectionKeys := make(map[string]struct{}, len(params.Collections))
	for i, c := range params.Collections {
		name, err := collection.ValidateName(c.Name)
		if err != nil {
			return ImportResult{}, fmt.Errorf("collections[%d]: %w", i, err)
		}
		slug, err := collection.ValidateSlug(c.Slug)
		if err != nil {
			return ImportResult{}, fmt.Errorf("collections[%d]: %w", i, err)
		}
		description, err := collection.ValidateDescription(c.Description)
		if err != nil {
			return ImportResult{}, fmt.Errorf("collections[%d]: %w", i, err)
		}

		key := strings.ToLower(name)
		if _, ok := collectionKeys[key]; ok {
			continue
		}
		collectionKeys[key] = struct{}{}
		collections = append(collections, importCollection{key: key, name: name, slug: slug, description: description})
	}

	// validate links data and collect unique tags
	now := time.Now()
	links := make([]importLink, len(params.Links))
	tagNames := []string{}
	tagSeen := map[string]struct{}{}
	for i, l := range params.Links {
		title, err := link.ValidateTitle(l.Title)
		if err != nil {
			return ImportResult{}, fmt.Errorf("links[%d]: %w", i, err)
		}
		description, err := link.ValidateDescription(l.Description)
		if err != nil {
			return ImportResult{}, fmt.Errorf("links[%d]: %w", i, err)
		}
		url, err := link.ValidateURL(l.URL)
		if err != nil {
			return ImportResult{}, fmt.Errorf("links[%d]: %w", i, err)
		}
		tags, err := tag.ValidateTagNames(l.Tags)
		if err != nil {
			return ImportResult{}, fmt.Errorf("links[%d]: %w", i, err)
		}

		var collectionKey string
		if l.Collection != nil {
			name, err := collection.ValidateName(*l.Collection)
			if err != nil {
				return ImportResult{}, fmt.Errorf("links[%d]: %w", i, err)
			}
			collectionKey = strings.ToLower(name)
			if _, ok := collectionKeys[collectionKey]; !ok {
				return ImportResult{}, fmt.Errorf("links[%d]: %w", i, ErrImportUnknownCollection)
			}
		}

		createdAt := now.Add(-time.Duration(i) * time.Microsecond)
		if l.CreatedAt != nil && l.CreatedAt.Before(now) {
			createdAt = *l.CreatedAt
		}

		for _, name := range tags {
			if _, ok := tagSeen[name]; !ok {
				tagSeen[name] = struct{}{}
				tagNames = append(tagNames, name)
			}
		}

		links[i] = importLink{
			title:         title,
			description:   description,
			url:           url,
			tags:          tags,
			favourite:     l.Favourite,
			forLater:      l.ForLater,
			createdAt:     createdAt,
			collectionKey: collectionKey,
		}
	}

	ctx, span := tracer.Start(ctx, "uow.Import")
	defer span.End()

	var result ImportResult

	err := s.uow.RunInTx(ctx, func(repos Repositories) error {
		// early return if import exceeds links limit
		if !params.Unlimited {
			count, err := repos.Links.Count(ctx, link.ListFilter{UserID: params.UserID})
			if err != nil {
				return err
			}
			if count+len(links) > config.FreeLinksLimit {
				return link.ErrLimitReached
			}
		}

		existingCollections, err := repos.Collections.List(ctx, params.UserID)
		if err != nil {
			return err
		}
		collectionIDs := make(map[string]uuid.UUID, len(existingCollections)+len(collections))
		for _, c := range existingCollections {
			collectionIDs[strings.ToLower(c.Name)] = c.ID
		}

		newCollections := []importCollection{}
		for _, c := range collections {
			if _, ok := collectionIDs[c.key]; !ok {
				newCollections = append(newCollections, c)
			}
		}

		// early return if import exceeds collections limit
		if !params.Unlimited && len(newCollections) > 0 && len(existingCollections)+len(newCollections) > config.FreeCollectionsLimit {
			return collection.ErrLimitReached
		}

		// create new collections and store their IDs
		for _, c := range newCollections {
			created, err := repos.Collections.Create(ctx, collection.CreateParams{
				UserID:      params.UserID,
				Name:        c.name,
				Slug:        c.slug,
				Description: c.description,
			})
			if err != nil {
				return err
			}
			collectionIDs[c.key] = created.ID
		}

		var tagsCountBefore, tagsCountAfter int
		if len(tagNames) > 0 {
			tagsCountBefore, err = repos.Tags.Count(ctx, params.UserID)
			if err != nil {
				return err
			}

			for _, name := range tagNames {
				if _, err := repos.Tags.UpsertByName(ctx, params.UserID, name); err != nil {
					return err
				}
			}

			tagsCountAfter, err = repos.Tags.Count(ctx, params.UserID)
			if err != nil {
				return err
			}

			// early return if import exceeds tags limit
			if !params.Unlimited && tagsCountAfter > tagsCountBefore && tagsCountAfter > config.FreeTagsLimit {
				return tag.ErrLimitReached
			}
		}

		// prebuild link.CreateManyParams slice to avoid multiple allocations
		newLinks := make([]link.CreateManyParams, len(links))
		for i, l := range links {
			var collectionID *uuid.UUID
			if l.collectionKey != "" {
				id := collectionIDs[l.collectionKey]
				collectionID = &id
			}

			newLinks[i] = link.CreateManyParams{
				Title:        l.title,
				Description:  l.description,
				URL:          l.url,
				CollectionID: collectionID,
				Tags:         l.tags,
				Favourite:    l.favourite,
				ForLater:     l.forLater,
				CreatedAt:    l.createdAt,
			}
		}

		if err := repos.Links.CreateMany(ctx, params.UserID, newLinks); err != nil {
			return err
		}

		result = ImportResult{
			Links:       len(newLinks),
			Collections: len(newCollections),
			Tags:        tagsCountAfter - tagsCountBefore,
		}

		return nil
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return result, err
}
