package user

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
	if errors.Is(err, ErrValidationUsernameRequired) ||
		errors.Is(err, ErrValidationUsernameTooShort) ||
		errors.Is(err, ErrValidationUsernameTooLong) ||
		errors.Is(err, ErrValidationUsernameCharacters) ||
		errors.Is(err, ErrValidationUsernamePrefix) ||
		errors.Is(err, ErrValidationUsernameSuffix) ||
		errors.Is(err, ErrValidationUsernameReserved) ||
		errors.Is(err, ErrValidationEmailRequired) ||
		errors.Is(err, ErrValidationEmailFormat) ||
		errors.Is(err, ErrValidationEmailTooLong) ||
		errors.Is(err, ErrValidationPasswordRequired) ||
		errors.Is(err, ErrValidationPasswordTooShort) ||
		errors.Is(err, ErrValidationPasswordTooLong) ||
		errors.Is(err, ErrValidationTokenRequired) ||
		errors.Is(err, ErrValidationTokenFormat) ||
		errors.Is(err, ErrValidationIsAdminRequired) ||
		errors.Is(err, ErrValidationIsProRequired) ||
		errors.Is(err, ErrValidationTokenDescriptionRequired) ||
		errors.Is(err, ErrValidationTokenDescriptionTooLong) {
		return http.StatusBadRequest, err.Error()
	}

	if errors.Is(err, ErrNotFound) {
		return http.StatusNotFound, "not found"
	}
	if errors.Is(err, ErrConflict) {
		return http.StatusConflict, "conflict"
	}
	if errors.Is(err, ErrInvalidCredentials) {
		return http.StatusUnauthorized, err.Error()
	}
	if errors.Is(err, ErrEmailNotVerified) {
		return http.StatusForbidden, "invalid email or password"
	}
	if errors.Is(err, ErrTokenExpired) {
		return http.StatusUnauthorized, err.Error()
	}
	if errors.Is(err, ErrResendTooFrequent) {
		return http.StatusTooManyRequests, err.Error()
	}
	if errors.Is(err, ErrPasswordResetTooFrequent) {
		return http.StatusTooManyRequests, err.Error()
	}

	log.Printf("Service error: %v", err)
	return http.StatusInternalServerError, "internal server error"
}
