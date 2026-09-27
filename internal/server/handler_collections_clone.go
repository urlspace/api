package server

import (
	"net/http"

	"github.com/urlspace/api/internal/collection"
	"uuid"
)

func handleCollectionsClone(svc *collection.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := getUserIDFromContext(r.Context())

		sourceID, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			writeJSONError(w, http.StatusNotFound, "not found")
			return
		}

		result, err := svc.Clone(r.Context(), sourceID, userID)
		if err != nil {
			statusCode, errorMessage := collection.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusCreated, collectionCreateResponse{
			Status: "ok",
			Data:   newResponseCollection(result),
		})
	}
}
