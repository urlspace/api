package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/zapi-sh/api/internal/db"
	"github.com/zapi-sh/api/internal/response"
	"github.com/zapi-sh/api/internal/store"
	"github.com/zapi-sh/api/internal/utils"
)

type ResourceCreateBody struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Favourite   *bool  `json:"favourite"`
	ReadLater   *bool  `json:"readLater"`
}

func (b *ResourceCreateBody) Validate() error {
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

type ResourceCreateResponse struct {
	Status string      `json:"status"`
	Data   db.Resource `json:"data"`
}

func ResourcesCreate(store *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body ResourceCreateBody
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

		rr, err := store.Resources.Create(r.Context(), body.Title, body.Description, body.URL, *body.Favourite, *body.ReadLater)
		if err != nil {
			response.HandleDbError(w, err)
			return
		}

		response := &ResourceCreateResponse{
			Status: "ok",
			Data:   rr,
		}

		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
