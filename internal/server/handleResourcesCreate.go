package server

import (
	"encoding/json"
	"net/http"

	"github.com/hreftools/api/internal/resource"
	"github.com/hreftools/api/internal/response"
	"github.com/hreftools/api/internal/utils"
	"github.com/hreftools/api/internal/validator"
)

type resourceCreateBody struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Favourite   *bool  `json:"favourite"`
	ReadLater   *bool  `json:"readLater"`
}

func (b *resourceCreateBody) validate() error {
	if err := validator.ResourceTitle(b.Title); err != nil {
		return err
	}

	if err := validator.Url(b.URL); err != nil {
		return err
	}

	if err := validator.ResourceDescription(b.Description); err != nil {
		return err
	}

	if err := validator.ResourceFavourite(b.Favourite); err != nil {
		return err
	}

	if err := validator.ResourceReadLater(b.ReadLater); err != nil {
		return err
	}

	return nil
}

type resourceCreateResponse struct {
	Status string           `json:"status"`
	Data   responseResource `json:"data"`
}

func handleResourcesCreate(svc *resource.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := utils.UserIDFromContext(r.Context())

		var body resourceCreateBody
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			response.HandleClientError(w, err, "invalid request body")
			return
		}

		if err := body.validate(); err != nil {
			response.HandleClientError(w, err, err.Error())
			return
		}

		params := resource.CreateParams{
			UserID:      userID,
			Title:       body.Title,
			Url:         body.URL,
			Description: body.Description,
			Favourite:   *body.Favourite,
			ReadLater:   *body.ReadLater,
		}
		rr, err := svc.Create(r.Context(), params)
		if err != nil {
			response.HandleDbError(w, err)
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
