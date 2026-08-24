package server

import (
	"net/http"

	"github.com/urlspace/api/internal/uow"
	"uuid"
)

type linkGetResponse struct {
	Status string       `json:"status"`
	Data   responseLink `json:"data"`
}

func handleLinksGet(uowSvc *uow.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := userIDFromContext(r.Context())

		id := r.PathValue("id")
		idUuid, err := uuid.Parse(id)
		if err != nil {
			handleClientError(r.Context(), w, err, "invalid id parameter")
			return
		}

		result, err := uowSvc.GetLink(r.Context(), idUuid, userID)
		if err != nil {
			statusCode, errorMessage := uow.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		writeJSONSuccess(w, http.StatusOK, linkGetResponse{
			Status: "ok",
			Data:   newResponseLink(result),
		})
	}
}
