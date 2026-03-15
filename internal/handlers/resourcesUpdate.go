package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/hreftools/api/internal/db"
	"github.com/hreftools/api/internal/response"
	"github.com/hreftools/api/internal/store"
	"github.com/hreftools/api/internal/utils"
	"github.com/hreftools/api/internal/validator"
)

type ResourceUpdateBody struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Favourite   *bool  `json:"favourite"`
	ReadLater   *bool  `json:"readLater"`
}

func (b *ResourceUpdateBody) Validate() error {
	if err := validator.ResourceTitle(b.Title); err != nil {
		return err
	}

	if err := validator.Url(b.URL); err != nil {
		return err
	}

	if err := validator.ResourceDescription(b.Description); err != nil {
		return err
	}

	if b.Favourite == nil {
		return errors.New("favourite field is required")
	}

	if b.ReadLater == nil {
		return errors.New("readLater field is required")
	}

	return nil
}

type ResourceUpdateResponse struct {
	Status string      `json:"status"`
	Data   db.Resource `json:"data"`
}

func ResourcesUpdate(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := utils.UserIDFromContext(r.Context())

		id := r.PathValue("id")
		idUuid, err := uuid.Parse(id)
		if err != nil {
			response.HandleClientError(w, err, "invalid id parameter")
			return
		}

		var body ResourceUpdateBody
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			response.HandleClientError(w, err, "invalid request body")
			return
		}

		if err := body.Validate(); err != nil {
			response.HandleClientError(w, err, err.Error())
			return
		}

		params := store.ResourceUpdateParams{
			ID:          idUuid,
			UserID:      userID,
			Title:       body.Title,
			Url:         body.URL,
			Description: body.Description,
			Favourite:   *body.Favourite,
			ReadLater:   *body.ReadLater,
		}
		rr, err := s.Resources.Update(r.Context(), params)
		if err != nil {
			response.HandleDbError(w, err)
			return
		}

		response := &ResourceUpdateResponse{
			Status: "ok",
			Data:   rr,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
