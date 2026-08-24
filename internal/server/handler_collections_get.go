package server

import (
	"net/http"

	"github.com/urlspace/api/internal/collection"
	"uuid"
)

type collectionsGetResponse struct {
	Status string             `json:"status"`
	Data   responseCollection `json:"data"`
}

func handleCollectionsGet(collectionSvc *collection.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := userIDFromContext(r.Context())

		id := r.PathValue("id")
		idUuid, err := uuid.Parse(id)
		if err != nil {
			handleClientError(r.Context(), w, err, "invalid id parameter")
			return
		}

		result, err := collectionSvc.Get(r.Context(), idUuid, userID)
		if err != nil {
			statusCode, errorMessage := collection.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusOK, collectionsGetResponse{
			Status: "ok",
			Data:   newResponseCollection(result),
		})
	}
}
