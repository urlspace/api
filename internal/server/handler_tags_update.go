package server

import (
	"encoding/json"
	"net/http"

	"github.com/urlspace/api/internal/tag"
	"uuid"
)

type tagUpdateBody struct {
	Name string `json:"name"`
}

type tagUpdateResponse struct {
	Status string      `json:"status"`
	Data   responseTag `json:"data"`
}

func handleTagsUpdate(svc *tag.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := userIDFromContext(r.Context())

		id := r.PathValue("id")
		tagID, err := uuid.Parse(id)
		if err != nil {
			handleClientError(r.Context(), w, err, "invalid id parameter")
			return
		}

		var body tagUpdateBody
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			handleClientError(r.Context(), w, err, "invalid request body")
			return
		}

		params := tag.UpdateParams{
			ID:     tagID,
			UserID: userID,
			Name:   body.Name,
		}
		t, err := svc.Update(r.Context(), params)
		if err != nil {
			statusCode, errorMessage := tag.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusOK, tagUpdateResponse{
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
