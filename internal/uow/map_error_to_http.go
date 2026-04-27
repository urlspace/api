package uow

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/urlspace/api/internal/collection"
	"github.com/urlspace/api/internal/link"
	"github.com/urlspace/api/internal/tag"
)

// MapErrorToHTTP maps errors from the uow service to HTTP status codes.
// This covers both link and tag validation errors since the uow service
// coordinates across both domains.
func MapErrorToHTTP(err error) (int, string) {
	if errors.Is(err, context.DeadlineExceeded) {
		return http.StatusRequestTimeout, "request timeout"
	}
	if errors.Is(err, context.Canceled) {
		return 499, "request cancelled"
	}

	// link validation errors
	if errors.Is(err, link.ErrValidationTitleLength) ||
		errors.Is(err, link.ErrValidationTitleInvalidCharacters) ||
		errors.Is(err, link.ErrValidationDescriptionLength) ||
		errors.Is(err, link.ErrValidationDescriptionInvalidCharacters) ||
		errors.Is(err, link.ErrValidationURLFormat) ||
		errors.Is(err, link.ErrValidationURLTooLong) ||
		errors.Is(err, link.ErrValidationURLPrivate) {
		return http.StatusBadRequest, err.Error()
	}

	// tag validation errors
	if errors.Is(err, tag.ErrValidationNameLength) ||
		errors.Is(err, tag.ErrValidationNameCharacters) ||
		errors.Is(err, tag.ErrValidationNameHyphens) ||
		errors.Is(err, tag.ErrValidationTooManyTags) {
		return http.StatusBadRequest, err.Error()
	}

	if errors.Is(err, link.ErrNotFound) || errors.Is(err, tag.ErrNotFound) || errors.Is(err, collection.ErrNotFound) {
		return http.StatusNotFound, "not found"
	}
	if errors.Is(err, link.ErrConflict) || errors.Is(err, tag.ErrConflict) || errors.Is(err, collection.ErrConflict) {
		return http.StatusConflict, "conflict"
	}

	log.Printf("Service error: %v", err)
	return http.StatusInternalServerError, "internal server error"
}
