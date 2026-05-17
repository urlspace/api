package server

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

// Decision (future me): must run after commonHeadersMiddleware (so the 500
// carries CORS/security headers) and before maxBody/auth/handlers (so it
// catches their panics). Without it, panics fall to net/http's default —
// stdlib log (no slog/trace context) and an abrupt connection close.
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.ErrorContext(r.Context(), "handler panicked",
					slog.Any("recover", rec),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("stack", string(debug.Stack())),
				)
				writeJSONError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
