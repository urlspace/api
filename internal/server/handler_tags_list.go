package server

import (
	"net/http"

	"github.com/urlspace/api/internal/tag"
)

type responseTagListItem struct {
	responseTag
	Count int `json:"count"`
}

type tagsListResponse struct {
	Status string                `json:"status"`
	Data   []responseTagListItem `json:"data"`
}

func handleTagsList(svc *tag.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := userIDFromContext(r.Context())

		list, err := svc.List(r.Context(), userID)
		if err != nil {
			statusCode, errorMessage := tag.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		items := make([]responseTagListItem, len(list))
		for i, item := range list {
			items[i] = responseTagListItem{
				responseTag: responseTag{
					ID:        item.ID,
					Name:      item.Name,
					CreatedAt: item.CreatedAt,
					UpdatedAt: item.UpdatedAt,
				},
				Count: item.LinkCount,
			}
		}

		writeJSONSuccess(w, http.StatusOK, tagsListResponse{
			Status: "ok",
			Data:   items,
		})
	}
}
