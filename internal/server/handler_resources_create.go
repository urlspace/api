package server

import (
	"encoding/json"
	"net/http"

	"github.com/urlspace/api/internal/resource"
)

type resourceCreateBody struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

type resourceCreateResponse struct {
	Status string           `json:"status"`
	Data   responseResource `json:"data"`
}

func handleResourcesCreate(svc *resource.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := userIDFromContext(r.Context())

		var body resourceCreateBody
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			handleClientError(w, err, "invalid request body")
			return
		}

		params := resource.CreateParamsService{
			UserID:      userID,
			Title:       body.Title,
			Url:         body.URL,
			Description: body.Description,
		}
		rr, err := svc.Create(r.Context(), params)
		if err != nil {
			statusCode, errorMessage := resource.MapErrorToHTTP(err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		res := &resourceCreateResponse{
			Status: "ok",
			Data:   newResponseResource(rr),
		}

		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
