package telemetry

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

// multiHandler fans a single slog record out to several handlers so the same
// log line can land in stdout (terminal + Railway) and the OTel bridge
// (Grafana) without callers having to log twice.
type multiHandler []slog.Handler

func (h multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h multiHandler) Handle(ctx context.Context, r slog.Record) error {
	var errs []error
	for _, handler := range h {
		if handler.Enabled(ctx, r.Level) {
			if err := handler.Handle(ctx, r.Clone()); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...)
}

func (h multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := make(multiHandler, len(h))
	for i, handler := range h {
		next[i] = handler.WithAttrs(attrs)
	}
	return next
}

func (h multiHandler) WithGroup(name string) slog.Handler {
	next := make(multiHandler, len(h))
	for i, handler := range h {
		next[i] = handler.WithGroup(name)
	}
	return next
}

// initLoggerProvider builds an OTel logger provider that exports via OTLP/HTTP.
// Endpoint, headers, etc. are picked up from the standard OTEL_* env vars.
func initLoggerProvider(ctx context.Context) (*sdklog.LoggerProvider, error) {
	exporter, err := otlploghttp.New(ctx)
	if err != nil {
		return nil, err
	}
	provider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
	)
	return provider, nil
}

// attachOtelLogger upgrades the slog default to fan out to both stdout and the
// OTel bridge, so logs reach Grafana without losing terminal/Railway output.
func attachOtelLogger(provider *sdklog.LoggerProvider) {
	otelHandler := otelslog.NewHandler("github.com/urlspace/api", otelslog.WithLoggerProvider(provider))
	slog.SetDefault(slog.New(multiHandler{slog.NewJSONHandler(os.Stdout, nil), otelHandler}))
}
