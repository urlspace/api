package billing

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

	if errors.Is(err, ErrAlreadyPro) {
		return http.StatusConflict, err.Error()
	}
	if errors.Is(err, ErrAdmin) {
		return http.StatusConflict, err.Error()
	}

	slog.ErrorContext(ctx, "service error", slog.String("error", err.Error()), slog.String("domain", "billing"))
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
	return http.StatusInternalServerError, "internal server error"
}
