package server

import (
	"log/slog"
	"net/http"
	"time"
)

type wrapperWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrapperWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &wrapperWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapped, r)

		// Map status code to log level so platforms (Railway, etc.) classify
		// 5xx as errors and 4xx as warnings instead of bucketing everything
		// the same.
		level := slog.LevelInfo
		switch {
		case wrapped.statusCode >= 500:
			level = slog.LevelError
		case wrapped.statusCode >= 400:
			level = slog.LevelWarn
		}

		slog.Log(r.Context(), level, "http request",
			slog.Int("status_code", wrapped.statusCode),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int64("duration_nano", time.Since(start).Nanoseconds()),
			slog.Int64("duration_micro", time.Since(start).Microseconds()),
			slog.Int64("duration_milli", time.Since(start).Milliseconds()),
		)
	})
}
