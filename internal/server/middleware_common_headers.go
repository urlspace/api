package server

import "net/http"

// Decisions for future me:
//
//   - No Vary: Origin / Vary: Cookie. Cache-Control: no-store already prevents
//     any cache from storing these responses, so there is no cache key for
//     Vary to influence. Revisit if no-store is ever weakened or removed.
//
//   - No CSRF Origin-check middleware. SameSite=Lax on the session cookie
//     blocks cross-site cookie-bearing POST/PUT/DELETE at the browser level,
//     and all JSON endpoints are non-simple under CORS so they require a
//     preflight that the Origin check below only approves for the site
//     (siteURL) and the signed-in app (dashboardURL). An
//     Origin-check middleware was prototyped and rejected because it broke
//     the Bruno-based REST workflow (which can't set a custom Origin).
func commonHeadersMiddleware(siteURL string, dashboardURL string) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Security headers
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
			w.Header().Set("Cache-Control", "no-store")
			w.Header().Set("Referrer-Policy", "no-referrer")

			// CORS
			origin := r.Header.Get("Origin")
			if origin == siteURL || origin == dashboardURL {
				// browser needs these on the preflight and actual response
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")

				// This is just for browsers, after quick check of allowed methods and headers,
				// the preflight can be terminated early without hitting the actual handler.
				if r.Method == http.MethodOptions {
					// browser needs these only on preflight, otherwise they are ignored, so we can skip them on the actual response
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
					w.Header().Set("Access-Control-Max-Age", "86400")
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}

			// Others
			w.Header().Set("Content-Type", "application/json")

			next.ServeHTTP(w, r)
		})
	}

}
