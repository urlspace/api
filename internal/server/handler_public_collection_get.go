package server

import (
	"errors"
	"net/http"
	"time"

	"github.com/urlspace/api/internal/collection"
	"uuid"
)

type responsePublicAuthor struct {
	DisplayName string `json:"displayName"`
	Username    string `json:"username"`
}

type responsePublicLink struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	URL         string    `json:"url"`
}

type responsePublicCollection struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	CreatedAt   time.Time            `json:"createdAt"`
	UpdatedAt   time.Time            `json:"updatedAt"`
	Author      responsePublicAuthor `json:"author"`
	Links       []responsePublicLink `json:"links"`
}

type publicCollectionGetResponse struct {
	Status string                   `json:"status"`
	Data   responsePublicCollection `json:"data"`
}

func handlePublicCollectionGet(collectionSvc *collection.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")

		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			handleNotFound(w, r)
			return
		}

		result, err := collectionSvc.GetPublic(r.Context(), id)
		if errors.Is(err, collection.ErrNotFound) {
			handleNotFound(w, r)
			return
		}
		if err != nil {
			handleServerError(r.Context(), w, err, "failed to get public collection")
			return
		}

		links := make([]responsePublicLink, len(result.Links))
		for i, link := range result.Links {
			links[i] = responsePublicLink{
				ID:          link.ID,
				Title:       link.Title,
				Description: link.Description,
				CreatedAt:   link.CreatedAt,
				URL:         link.URL,
			}
		}

		writeJSONSuccess(w, http.StatusOK, publicCollectionGetResponse{
			Status: "ok",
			Data: responsePublicCollection{
				Name:        result.Name,
				Description: result.Description,
				CreatedAt:   result.CreatedAt,
				UpdatedAt:   result.UpdatedAt,
				Author: responsePublicAuthor{
					DisplayName: result.Author.DisplayName,
					Username:    result.Author.Username,
				},
				Links: links,
			},
		})
	}
}
