package server

import (
	"net/http"
	"os"
	"time"

	"github.com/zapi-sh/api/internal/handlers"
	"github.com/zapi-sh/api/internal/middlewares"
	"github.com/zapi-sh/api/internal/store"
)

func New(store *store.Store) *http.Server {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	// routes
	mux := http.NewServeMux()

	// not found
	mux.HandleFunc("/", handlers.NotFound)
	// status
	mux.HandleFunc("GET /status", handlers.Status)
	// resources
	mux.HandleFunc("GET /resources", handlers.ResourcesList(store))
	mux.HandleFunc("GET /resources/{id}", handlers.ResourcesGet(store))
	mux.HandleFunc("POST /resources", handlers.ResourcesCreate(store))
	mux.HandleFunc("PUT /resources/{id}", handlers.ResourcesUpdate(store))
	mux.HandleFunc("DELETE /resources/{id}", handlers.ResourcesDelete(store))

	// version api
	v1 := http.NewServeMux()
	v1.Handle("/v1/", http.StripPrefix("/v1", mux))

	// apply middlewares
	middlewaresStack := middlewares.MiddlewareStac(
		middlewares.Logging,
		middlewares.CommonHeaders,
	)

	return &http.Server{
		Addr:              ":" + port,
		Handler:           middlewaresStack(v1),
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

}
