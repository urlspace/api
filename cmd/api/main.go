package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/zapi-sh/api/internal/handlers"
	"github.com/zapi-sh/api/internal/middlewares"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /status", handlers.Status)

	handler := middlewares.CommonHeaders(mux)

	s := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	log.Printf("Starting server on %s", port)

	log.Fatal(s.ListenAndServe())
}
