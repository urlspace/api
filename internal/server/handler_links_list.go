package server

import (
	"fmt"
	"net/http"
	"strconv"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/urlspace/api/internal/uow"
)

const (
	linksListDefaultPageSize = 250
	linksListMaxPageSize     = 500
	// linksListMaxQueryLength matches the max title length. A query longer
	// than any title that could exist cannot match anything, so it's malformed
	// by definition. When description matching is added, bump to the larger of
	// title (255) and description (512) max lengths.
	linksListMaxQueryLength = 255
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
		q := r.URL.Query()

		page := 1
		pageSize := linksListDefaultPageSize

		if v := q.Get("page"); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil || n < 1 {
				writeJSONError(w, http.StatusBadRequest, "page must be a positive integer")
				return
			}
			page = n
		}

		if v := q.Get("pageSize"); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil || n < 1 || n > linksListMaxPageSize {
				writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("pageSize must be between 1 and %d", linksListMaxPageSize))
				return
			}
			pageSize = n
		}

		var collectionID *uuid.UUID
		if v := q.Get("collectionId"); v != "" {
			id, err := uuid.Parse(v)
			if err != nil {
				writeJSONError(w, http.StatusBadRequest, "collectionId must be a valid uuid")
				return
			}
			collectionID = &id
		}

		var tagIDs []uuid.UUID
		for _, v := range q["tagId"] {
			id, err := uuid.Parse(v)
			if err != nil {
				writeJSONError(w, http.StatusBadRequest, "tagId must be a valid uuid")
				return
			}
			tagIDs = append(tagIDs, id)
		}

		var favourite *bool
		if v := q.Get("favourite"); v != "" {
			b, err := strconv.ParseBool(v)
			if err != nil {
				writeJSONError(w, http.StatusBadRequest, "favourite must be true or false")
				return
			}
			favourite = &b
		}

		var forLater *bool
		if v := q.Get("forLater"); v != "" {
			b, err := strconv.ParseBool(v)
			if err != nil {
				writeJSONError(w, http.StatusBadRequest, "forLater must be true or false")
				return
			}
			forLater = &b
		}

		query := q.Get("query")
		if utf8.RuneCountInString(query) > linksListMaxQueryLength {
			writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("query must be at most %d characters", linksListMaxQueryLength))
			return
		}

		result, err := uowSvc.ListLinks(r.Context(), uow.ListLinksParams{
			UserID:       userID,
			Page:         page,
			PageSize:     pageSize,
			CollectionID: collectionID,
			TagIDs:       tagIDs,
			Query:        query,
			Favourite:    favourite,
			ForLater:     forLater,
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
