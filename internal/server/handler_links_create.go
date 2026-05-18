package server

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/urlspace/api/internal/uow"
)

type linkCreateBody struct {
	Title        string   `json:"title"`
	URL          string   `json:"url"`
	Description  string   `json:"description"`
	CollectionID *string  `json:"collectionId"`
	Tags         []string `json:"tags"`
	Favourite    bool     `json:"favourite"`
	ForLater     bool     `json:"forLater"`
}

type linkCreateResponse struct {
	Status string       `json:"status"`
	Data   responseLink `json:"data"`
}

func handleLinksCreate(uowSvc *uow.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := userIDFromContext(r.Context())

		var body linkCreateBody
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			handleClientError(r.Context(), w, err, "invalid request body")
			return
		}

		var collectionID *uuid.UUID
		if body.CollectionID != nil {
			id, err := uuid.Parse(*body.CollectionID)
			if err != nil {
				handleClientError(r.Context(), w, err, "invalid collectionId")
				return
			}
			collectionID = &id
		}

		result, err := uowSvc.CreateLink(r.Context(), uow.CreateLinkParams{
			UserID:       userID,
			Title:        body.Title,
			URL:          body.URL,
			Description:  body.Description,
			CollectionID: collectionID,
			Tags:         body.Tags,
			Favourite:    body.Favourite,
			ForLater:     body.ForLater,
		})
		if err != nil {
			statusCode, errorMessage := uow.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusCreated, linkCreateResponse{
			Status: "ok",
			Data:   newResponseLink(result),
		})
	}
}
