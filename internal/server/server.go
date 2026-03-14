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

	auth := middlewares.Auth(s)
	adminOnly := middlewares.MiddlewareStack(auth, middlewares.Admin(s))

	// resources (protected)
	mux.Handle("GET /resources", auth(handlers.ResourcesList(s)))
	mux.Handle("GET /resources/{id}", auth(handlers.ResourcesGet(s)))
	mux.Handle("POST /resources", auth(handlers.ResourcesCreate(s)))
	mux.Handle("PUT /resources/{id}", auth(handlers.ResourcesUpdate(s)))
	mux.Handle("DELETE /resources/{id}", auth(handlers.ResourcesDelete(s)))

	// users (admin only)
	mux.Handle("GET /users", adminOnly(handlers.UsersList(s)))
	mux.Handle("GET /users/{id}", adminOnly(handlers.UsersGet(s)))
	mux.Handle("POST /users", adminOnly(handlers.UserCreate(s)))
	mux.Handle("DELETE /users/{id}", adminOnly(handlers.UsersDelete(s)))

	// auth
	mux.HandleFunc("POST /auth/signup", handlers.AuthSignup(s, emailSender))
	mux.HandleFunc("POST /auth/signin", handlers.AuthSignin(s))
	mux.HandleFunc("POST /auth/verify", handlers.AuthVerify(s))
	mux.HandleFunc("POST /auth/resend-verification", handlers.AuthResendVerification(s, emailSender))
	mux.Handle("POST /auth/signout", auth(handlers.AuthSignout(s)))

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
