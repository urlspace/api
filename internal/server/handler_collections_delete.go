package server

import (
	"net/http"
	"strconv"

	"github.com/urlspace/api/internal/collection"
	"uuid"
)

type collectionDeleteResponse struct {
	Status string             `json:"status"`
	Data   responseCollection `json:"data"`
}

func handleCollectionsDelete(collectionSvc *collection.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := getUserIDFromContext(r.Context())

		id := r.PathValue("id")
		idUuid, err := uuid.Parse(id)
		if err != nil {
			handleClientError(r.Context(), w, err, "invalid id parameter")
			return
		}

		cascade := false
		if v := r.URL.Query().Get("cascade"); v != "" {
			b, err := strconv.ParseBool(v)
			if err != nil {
				writeJSONError(w, http.StatusBadRequest, "cascade must be true or false")
				return
			}
			cascade = b
		}

		result, err := collectionSvc.Delete(r.Context(), idUuid, userID, cascade)
		if err != nil {
			statusCode, errorMessage := collection.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusOK, collectionDeleteResponse{
			Status: "ok",
			Data:   newResponseCollection(result, canPublishFromContext(r.Context())),
		})
	}
}
