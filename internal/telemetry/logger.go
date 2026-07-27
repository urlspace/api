package telemetry

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelslog"
)

// InitLogging installs the default slog logger: a fan-out handler that writes
// JSON to stdout (terminal + Railway) and bridges records into the OTel logs
// pipeline (Grafana).
//
// There is no runtime OTel SDK setup here. When built with otelc (Docker/prod
// and `make run`), otelc's compile-time instrumentation creates the global
// tracer/meter/logger providers from OTEL_* env vars, and otelslog defaults to
// that global logger provider. In a plain build (the air `make dev` loop) the
// global provider is a no-op, so only stdout logging is active.
//
// Decision (future me): otelc's slog instrumentation only injects
// trace_id/span_id attributes; it does NOT ship slog records to the OTLP logs
// exporter, so this bridge is what actually delivers logs to Grafana.
func InitLogging() {
	otelHandler := otelslog.NewHandler("github.com/urlspace/api")
	slog.SetDefault(slog.New(multiHandler{
		slog.NewJSONHandler(os.Stdout, nil),
		otelHandler,
	}))
}

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
