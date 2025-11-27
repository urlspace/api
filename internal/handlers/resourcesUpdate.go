package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/zapi-sh/api/internal/db"
	"github.com/zapi-sh/api/internal/response"
	"github.com/zapi-sh/api/internal/store"
	"github.com/zapi-sh/api/internal/utils"
)

type ResourceUpdateBody struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Favourite   *bool  `json:"favourite"`
	ReadLater   *bool  `json:"readLater"`
}

func (b *ResourceUpdateBody) Validate() error {
	if len(b.Title) < store.ResourceTitleLengthMin || len(b.Title) > store.ResourceTitleLengthMax {
		return errors.New("title must be between 3 and 255 characters")
	}

	if utils.IsValidURL(b.URL) == false {
		return errors.New("url must be a valid URL")
	}

	if len(b.Description) > store.ResourceDescriptionLengthMax {
		return errors.New("description must be less than 512 characters")
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

func ResourcesUpdate(store *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		rr, err := store.Resources.Update(r.Context(), idUuid, body.Title, body.Description, body.URL, *body.Favourite, *body.ReadLater)
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
