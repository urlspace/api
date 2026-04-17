package resource

import (
	"context"
	"errors"
	"log"
	"net/http"
)

func MapErrorToHTTP(err error) (int, string) {
	// context errors
	if errors.Is(err, context.DeadlineExceeded) {
		return http.StatusRequestTimeout, "request timeout"
	}
	if errors.Is(err, context.Canceled) {
		return 499, "request cancelled"
	}

	// validation errors
	if errors.Is(err, ErrValidationTitleLength) ||
		errors.Is(err, ErrValidationTitleInvalidCharacters) ||
		errors.Is(err, ErrValidationDescriptionLength) ||
		errors.Is(err, ErrValidationDescriptionInvalidCharacters) ||
		errors.Is(err, ErrValidationURLFormat) ||
		errors.Is(err, ErrValidationURLTooLong) ||
		errors.Is(err, ErrValidationURLPrivate) ||
		errors.Is(err, ErrValidationFavouriteRequired) ||
		errors.Is(err, ErrValidationReadLaterRequired) {
		return http.StatusBadRequest, err.Error()
	}

	if errors.Is(err, ErrNotFound) {
		return http.StatusNotFound, "not found"
	}
	if errors.Is(err, ErrConflict) {
		return http.StatusConflict, "conflict"
	}

	log.Printf("Service error: %v", err)
	return http.StatusInternalServerError, "internal server error"
}
