package server

import (
	"encoding/json/v2"
	"net/http"

	"github.com/urlspace/api/internal/collection"
	"uuid"
)

type collectionUpdateBody struct {
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description string  `json:"description"`
	Public      bool    `json:"public"`
	Password    *string `json:"password"`
}

type collectionUpdateResponse struct {
	Status string             `json:"status"`
	Data   responseCollection `json:"data"`
}

func handleCollectionsUpdate(collectionSvc *collection.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := getUserIDFromContext(r.Context())

		id := r.PathValue("id")
		idUuid, err := uuid.Parse(id)
		if err != nil {
			handleClientError(r.Context(), w, err, "invalid id parameter")
			return
		}

		var body collectionUpdateBody
		if err := json.UnmarshalRead(r.Body, &body, json.RejectUnknownMembers(true)); err != nil {
			handleClientError(r.Context(), w, err, "invalid request body")
			return
		}

		canPublish := canPublishFromContext(r.Context())
		if body.Public && !canPublish {
			writeJSONError(w, http.StatusForbidden, "forbidden")
			return
		}

		// Users without Unlimited can't change the flag, so keep the stored value
		// instead of the false the client sends. Upgrading again restores public
		// pages.
		public := body.Public
		slug := body.Slug
		password := body.Password
		if !canPublish {
			current, err := collectionSvc.Get(r.Context(), idUuid, userID)
			if err != nil {
				statusCode, errorMessage := collection.MapErrorToHTTP(r.Context(), err)
				writeJSONError(w, statusCode, errorMessage)
				return
			}
			public = current.Public
			// A collection that was public keeps its slug and password, so its
			// address and access are the same when the user upgrades again.
			if current.Public {
				slug = current.Slug
				password = current.Password
			}
		}

		result, err := collectionSvc.Update(r.Context(), collection.UpdateParams{
			ID:          idUuid,
			UserID:      userID,
			Name:        body.Name,
			Slug:        slug,
			Description: body.Description,
			Public:      public,
			Password:    password,
		})
		if err != nil {
			statusCode, errorMessage := collection.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusOK, collectionUpdateResponse{
			Status: "ok",
			Data:   newResponseCollection(result, canPublish),
		})
	}
}
