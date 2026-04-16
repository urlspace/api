package server

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/hreftools/api/internal/resource"
)

type resourceUpdateBody struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Favourite   *bool  `json:"favourite"`
	ReadLater   *bool  `json:"readLater"`
}

type resourceUpdateResponse struct {
	Status string           `json:"status"`
	Data   responseResource `json:"data"`
}

func handleResourcesUpdate(svc *resource.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := userIDFromContext(r.Context())

		id := r.PathValue("id")
		idUuid, err := uuid.Parse(id)
		if err != nil {
			handleClientError(w, err, "invalid id parameter")
			return
		}

		var body resourceUpdateBody
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			handleClientError(w, err, "invalid request body")
			return
		}

		params := resource.UpdateParamsService{
			ID:          idUuid,
			UserID:      userID,
			Title:       body.Title,
			Url:         body.URL,
			Description: body.Description,
			Favourite:   body.Favourite,
			ReadLater:   body.ReadLater,
		}
		rr, err := svc.Update(r.Context(), params)
		if err != nil {
			statusCode, errorMessage := resource.MapErrorToHTTP(err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		res := &resourceUpdateResponse{
			Status: "ok",
			Data:   newResponseResource(rr),
		}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
