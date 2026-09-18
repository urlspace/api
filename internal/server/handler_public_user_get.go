package server

import (
	"errors"
	"net/http"
	"time"

	"github.com/urlspace/api/internal/user"
	"uuid"
)

type responsePublicUserCollection struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type responsePublicUser struct {
	DisplayName string                         `json:"displayName"`
	Collections []responsePublicUserCollection `json:"collections"`
}

type publicUserGetResponse struct {
	Status string             `json:"status"`
	Data   responsePublicUser `json:"data"`
}

func handlePublicUserGet(userSvc *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := userSvc.GetPublic(r.Context(), r.PathValue("username"))
		if errors.Is(err, user.ErrNotFound) {
			handleNotFound(w, r)
			return
		}
		if err != nil {
			handleServerError(r.Context(), w, err, "failed to get public user")
			return
		}

		collections := make([]responsePublicUserCollection, len(result.Collections))
		for i, collection := range result.Collections {
			collections[i] = responsePublicUserCollection{
				ID:          collection.ID,
				Name:        collection.Name,
				Description: collection.Description,
				CreatedAt:   collection.CreatedAt,
				UpdatedAt:   collection.UpdatedAt,
			}
		}

		writeJSONSuccess(w, http.StatusOK, publicUserGetResponse{
			Status: "ok",
			Data: responsePublicUser{
				DisplayName: result.DisplayName,
				Collections: collections,
			},
		})
	}
}
