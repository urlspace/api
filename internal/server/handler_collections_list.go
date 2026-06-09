package server

import (
	"net/http"

	"github.com/urlspace/api/internal/collection"
)

type responseCollectionListItem struct {
	responseCollection
	Count int `json:"count"`
}

type collectionsListResponse struct {
	Status string                       `json:"status"`
	Data   []responseCollectionListItem `json:"data"`
}

func handleCollectionsList(collectionSvc *collection.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := userIDFromContext(r.Context())

		list, err := collectionSvc.List(r.Context(), userID)
		if err != nil {
			statusCode, errorMessage := collection.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		items := make([]responseCollectionListItem, len(list))
		for i, item := range list {
			items[i] = responseCollectionListItem{
				responseCollection: newResponseCollection(item.Collection),
				Count:              item.LinkCount,
			}
		}

		writeJSONSuccess(w, http.StatusOK, collectionsListResponse{
			Status: "ok",
			Data:   items,
		})
	}
}
