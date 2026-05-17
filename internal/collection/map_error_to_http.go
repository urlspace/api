package collection

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func MapErrorToHTTP(ctx context.Context, err error) (int, string) {
	// context errors
	if errors.Is(err, context.DeadlineExceeded) {
		return http.StatusRequestTimeout, "request timeout"
	}
	if errors.Is(err, context.Canceled) {
		return 499, "request cancelled"
	}

	// validation errors
	if errors.Is(err, ErrValidationNameLength) ||
		errors.Is(err, ErrValidationNameInvalidCharacters) ||
		errors.Is(err, ErrValidationDescriptionLength) ||
		errors.Is(err, ErrValidationDescriptionInvalidCharacters) {
		return http.StatusBadRequest, err.Error()
	}

	if errors.Is(err, ErrNotFound) {
		return http.StatusNotFound, "not found"
	}
	if errors.Is(err, ErrConflict) {
		return http.StatusConflict, "conflict"
	}

	slog.ErrorContext(ctx, "service error", slog.String("error", err.Error()), slog.String("domain", "collection"))
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
	return http.StatusInternalServerError, "internal server error"
}
