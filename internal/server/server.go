package server

import (
	"github.com/zapi-sh/api/internal/handlers"
	"github.com/zapi-sh/api/internal/middlewares"
	"net/http"
	"os"
	"time"
)

func New() *http.Server {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /status", handlers.Status)

	handler := middlewares.CommonHeaders(mux)

	return &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

}
