package server

import (
	"encoding/json/v2"
	"net/http"

	"github.com/urlspace/api/internal/collection"
	"uuid"
)

type collectionCloneBody struct {
	Name string `json:"name"`
}

func handleCollectionsClone(svc *collection.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := getUserIDFromContext(r.Context())

		sourceID, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			writeJSONError(w, http.StatusNotFound, "not found")
			return
		}

		var body collectionCloneBody
		if err := json.UnmarshalRead(r.Body, &body, json.RejectUnknownMembers(true)); err != nil {
			handleClientError(r.Context(), w, err, "invalid request body")
			return
		}

		result, err := svc.Clone(r.Context(), sourceID, userID, body.Name, isUnlimitedFromContext(r.Context()))
		if err != nil {
			statusCode, errorMessage := collection.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusCreated, collectionCreateResponse{
			Status: "ok",
			Data:   newResponseCollection(result, canPublishFromContext(r.Context())),
		})
	}
}
