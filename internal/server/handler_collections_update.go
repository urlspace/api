package server

import (
	"encoding/json"
	"net/http"

	"github.com/urlspace/api/internal/collection"
	"uuid"
)

type collectionUpdateBody struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Public      bool   `json:"public"`
}

type collectionUpdateResponse struct {
	Status string             `json:"status"`
	Data   responseCollection `json:"data"`
}

func handleCollectionsUpdate(collectionSvc *collection.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := userIDFromContext(r.Context())

		id := r.PathValue("id")
		idUuid, err := uuid.Parse(id)
		if err != nil {
			handleClientError(r.Context(), w, err, "invalid id parameter")
			return
		}

		var body collectionUpdateBody
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			handleClientError(r.Context(), w, err, "invalid request body")
			return
		}

		result, err := collectionSvc.Update(r.Context(), collection.UpdateParams{
			ID:          idUuid,
			UserID:      userID,
			Name:        body.Name,
			Description: body.Description,
			Public:      body.Public,
		})
		if err != nil {
			statusCode, errorMessage := collection.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusOK, collectionUpdateResponse{
			Status: "ok",
			Data:   newResponseCollection(result),
		})
	}
}
