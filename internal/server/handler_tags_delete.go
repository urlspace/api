package server

import (
	"net/http"

	"github.com/urlspace/api/internal/tag"
	"uuid"
)

type tagDeleteResponse struct {
	Status string      `json:"status"`
	Data   responseTag `json:"data"`
}

func handleTagsDelete(svc *tag.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := userIDFromContext(r.Context())

		id := r.PathValue("id")
		tagID, err := uuid.Parse(id)
		if err != nil {
			handleClientError(r.Context(), w, err, "invalid id parameter")
			return
		}

		t, err := svc.Delete(r.Context(), tagID, userID)
		if err != nil {
			statusCode, errorMessage := tag.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusOK, tagDeleteResponse{
			Status: "ok",
			Data: responseTag{
				ID:        t.ID,
				Name:      t.Name,
				CreatedAt: t.CreatedAt,
				UpdatedAt: t.UpdatedAt,
			},
		})
	}
}
