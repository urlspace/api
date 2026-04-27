package server

import (
	"net/http"

	"github.com/urlspace/api/internal/uow"
)

type linksListResponse struct {
	Status string         `json:"status"`
	Data   []responseLink `json:"data"`
}

func handleLinksList(uowSvc *uow.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := userIDFromContext(r.Context())

		list, err := uowSvc.ListLinks(r.Context(), userID)
		if err != nil {
			statusCode, errorMessage := uow.MapErrorToHTTP(err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		items := make([]responseLink, len(list))
		for i, item := range list {
			items[i] = newResponseLink(item)
		}

		writeJSONSuccess(w, http.StatusOK, linksListResponse{
			Status: "ok",
			Data:   items,
		})
	}
}
