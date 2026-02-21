package server

import (
	"net/http"
	"os"
	"time"

	"github.com/hreftools/api/internal/emails"
	"github.com/hreftools/api/internal/handlers"
	"github.com/hreftools/api/internal/middlewares"
	"github.com/hreftools/api/internal/store"
)

func New(s *store.Store, emailSender emails.EmailSender) *http.Server {
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
	// this requires an authentication, middlware to come later here
	mux.HandleFunc("GET /resources", handlers.ResourcesList(s))
	mux.HandleFunc("GET /resources/{id}", handlers.ResourcesGet(s))
	mux.HandleFunc("POST /resources", handlers.ResourcesCreate(s))
	mux.HandleFunc("PUT /resources/{id}", handlers.ResourcesUpdate(s))
	mux.HandleFunc("DELETE /resources/{id}", handlers.ResourcesDelete(s))

	// users
	// this is for admins only, middleware to come later here
	mux.HandleFunc("GET /users", handlers.UsersList(s))
	mux.HandleFunc("GET /users/{id}", handlers.UsersGet(s))
	mux.HandleFunc("POST /users", handlers.UserCreate(s))
	mux.HandleFunc("DELETE /users/{id}", handlers.UsersDelete(s))

	// auth
	mux.HandleFunc("POST /auth/signup", handlers.AuthSignup(s, emailSender))
	// mux.HandleFunc("POST /auth/signin", handlers.xxx(store))
	// mux.HandleFunc("POST /auth/signout", handlers.xxx(store))
	mux.HandleFunc("POST /auth/verify", handlers.AuthVerify(s))
	mux.HandleFunc("POST /auth/resend-verification", handlers.AuthResendVerification(s, emailSender))

	// version api
	v1 := http.NewServeMux()
	v1.Handle("/v1/", http.StripPrefix("/v1", mux))

	// apply middlewares
	middlewaresStack := middlewares.MiddlewareStack(
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
