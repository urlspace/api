package server

import "net/http"

func maxBodySizeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit := int64(64 << 10) // 64 KB
		if r.Method == http.MethodPost && r.URL.Path == "/v1/tools/import" {
			limit = 32 << 20 // 32 MB
		}
		r.Body = http.MaxBytesReader(w, r.Body, limit)
		next.ServeHTTP(w, r)
	})
}
