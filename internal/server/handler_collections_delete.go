package server

import (
	"net/http"

	"github.com/urlspace/api/internal/collection"
	"uuid"
)

type collectionDeleteResponse struct {
	Status string             `json:"status"`
	Data   responseCollection `json:"data"`
}

func handleCollectionsDelete(collectionSvc *collection.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := userIDFromContext(r.Context())

		id := r.PathValue("id")
		idUuid, err := uuid.Parse(id)
		if err != nil {
			handleClientError(r.Context(), w, err, "invalid id parameter")
			return
		}

		result, err := collectionSvc.Delete(r.Context(), idUuid, userID)
		if err != nil {
			statusCode, errorMessage := collection.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusOK, collectionDeleteResponse{
			Status: "ok",
			Data:   newResponseCollection(result),
		})
	}
}
