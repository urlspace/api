package server

import (
	"encoding/json/v2"
	"net/http"
	"time"

	"github.com/urlspace/api/internal/uow"
)

type toolsImportCollection struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

type toolsImportLink struct {
	URL         string     `json:"url"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Tags        []string   `json:"tags"`
	Favourite   bool       `json:"favourite"`
	ForLater    bool       `json:"forLater"`
	CreatedAt   *time.Time `json:"createdAt"`
	Collection  *string    `json:"collection"`
}

type toolsImportBody struct {
	Collections []toolsImportCollection `json:"collections"`
	Links       []toolsImportLink       `json:"links"`
}

type responseImport struct {
	Links       int `json:"links"`
	Collections int `json:"collections"`
	Tags        int `json:"tags"`
}

type toolsImportResponse struct {
	Status string         `json:"status"`
	Data   responseImport `json:"data"`
}

func handleToolsImport(uowSvc *uow.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := getUserIDFromContext(r.Context())

		var body toolsImportBody
		if err := json.UnmarshalRead(r.Body, &body, json.RejectUnknownMembers(true)); err != nil {
			handleClientError(r.Context(), w, err, "invalid request body")
			return
		}

		collections := make([]uow.ImportCollection, len(body.Collections))
		for i, c := range body.Collections {
			collections[i] = uow.ImportCollection{
				Name:        c.Name,
				Slug:        c.Slug,
				Description: c.Description,
			}
		}

		links := make([]uow.ImportLink, len(body.Links))
		for i, l := range body.Links {
			links[i] = uow.ImportLink{
				Title:       l.Title,
				Description: l.Description,
				URL:         l.URL,
				Tags:        l.Tags,
				Favourite:   l.Favourite,
				ForLater:    l.ForLater,
				CreatedAt:   l.CreatedAt,
				Collection:  l.Collection,
			}
		}

		result, err := uowSvc.Import(r.Context(), uow.ImportParams{
			UserID:      userID,
			Collections: collections,
			Links:       links,
			Unlimited:   isUnlimitedFromContext(r.Context()),
		})
		if err != nil {
			statusCode, errorMessage := uow.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusCreated, toolsImportResponse{
			Status: "ok",
			Data: responseImport{
				Links:       result.Links,
				Collections: result.Collections,
				Tags:        result.Tags,
			},
		})
	}
}
