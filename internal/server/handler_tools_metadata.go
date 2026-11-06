package server

import (
	"encoding/json/v2"
	"net/http"

	"github.com/urlspace/api/internal/link"
)

type toolsMetadataBody struct {
	URL string `json:"url"`
}

type responseToolsMetadata struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type toolsMetadataResponse struct {
	Status string                `json:"status"`
	Data   responseToolsMetadata `json:"data"`
}

func handleToolsMetadata(linkSvc *link.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := getUserIDFromContext(r.Context())

		var body toolsMetadataBody
		if err := json.UnmarshalRead(r.Body, &body, json.RejectUnknownMembers(true)); err != nil {
			handleClientError(r.Context(), w, err, "invalid request body")
			return
		}

		u, err := link.ValidateURL(body.URL)
		if err != nil {
			handleClientError(r.Context(), w, err, "invalid url")
			return
		}

		metadata := linkSvc.FetchMetadata(r.Context(), userID, u, r.Header.Get("Accept-Language"))

		writeJSONSuccess(w, http.StatusOK, toolsMetadataResponse{
			Status: "ok",
			Data: responseToolsMetadata{
				Title:       metadata.Title,
				Description: metadata.Description,
			},
		})
	}
}
