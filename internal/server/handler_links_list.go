package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/urlspace/api/internal/uow"
)

const (
	linksListDefaultPageSize = 250
	linksListMaxPageSize     = 500
)

type responsePagination struct {
	TotalCount  int `json:"totalCount"`
	PageSize    int `json:"pageSize"`
	CurrentPage int `json:"currentPage"`
	TotalPages  int `json:"totalPages"`
}

type linksListResponse struct {
	Status     string             `json:"status"`
	Data       []responseLink     `json:"data"`
	Pagination responsePagination `json:"pagination"`
}

func handleLinksList(uowSvc *uow.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := userIDFromContext(r.Context())

		page := 1
		pageSize := linksListDefaultPageSize

		if v := r.URL.Query().Get("page"); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil || n < 1 {
				writeJSONError(w, http.StatusBadRequest, "page must be a positive integer")
				return
			}
			page = n
		}

		if v := r.URL.Query().Get("pageSize"); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil || n < 1 || n > linksListMaxPageSize {
				writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("pageSize must be between 1 and %d", linksListMaxPageSize))
				return
			}
			pageSize = n
		}

		result, err := uowSvc.ListLinks(r.Context(), uow.ListLinksParams{
			UserID:   userID,
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			statusCode, errorMessage := uow.MapErrorToHTTP(r.Context(), err)
			writeJSONError(w, statusCode, errorMessage)
			return
		}

		items := make([]responseLink, len(result.Links))
		for i, item := range result.Links {
			items[i] = newResponseLink(item)
		}

		// Integer ceil division: (a + b - 1) / b. Go's "/" truncates, so 319/100
		// is 3 not 3.19, which would give totalPages=3 for 319 links — one short.
		// Adding (pageSize-1) before dividing pushes partial pages up to the
		// next multiple.
		//
		// Examples (pageSize=100):
		//   totalCount=0   → (0 + 99) / 100   = 0
		//   totalCount=1   → (1 + 99) / 100   = 1
		//   totalCount=100 → (100 + 99) / 100 = 1
		//   totalCount=300 → (300 + 99) / 100 = 3
		//   totalCount=319 → (319 + 99) / 100 = 4
		totalPages := (result.TotalCount + pageSize - 1) / pageSize

		writeJSONSuccess(w, http.StatusOK, linksListResponse{
			Status: "ok",
			Data:   items,
			Pagination: responsePagination{
				TotalCount:  result.TotalCount,
				PageSize:    pageSize,
				CurrentPage: page,
				TotalPages:  totalPages,
			},
		})
	}
}
