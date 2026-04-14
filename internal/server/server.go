package server

import (
	"net/http"
	"time"

	"github.com/hreftools/api/internal/handlers"
	"github.com/hreftools/api/internal/middlewares"
	"github.com/hreftools/api/internal/resource"
	"github.com/hreftools/api/internal/user"
)

func New(port string, userSvc *user.Service, resourceSvc *resource.Service) *http.Server {
	// routes
	mux := http.NewServeMux()

	// not found
	mux.HandleFunc("/", handlers.NotFound)

	// status
	mux.HandleFunc("GET /status", handlers.Status)

	auth := middlewares.Auth(userSvc)
	adminOnly := middlewares.MiddlewareStack(auth, middlewares.Admin(userSvc))

	// me (authenticated)
	mux.Handle("GET /me", auth(handlers.MeGet(userSvc)))

	// resources (protected)
	mux.Handle("GET /resources", auth(handlers.ResourcesList(resourceSvc)))
	mux.Handle("GET /resources/{id}", auth(handlers.ResourcesGet(resourceSvc)))
	mux.Handle("POST /resources", auth(handlers.ResourcesCreate(resourceSvc)))
	mux.Handle("PUT /resources/{id}", auth(handlers.ResourcesUpdate(resourceSvc)))
	mux.Handle("DELETE /resources/{id}", auth(handlers.ResourcesDelete(resourceSvc)))

	// users (admin only)
	mux.Handle("GET /users", adminOnly(handlers.UsersList(userSvc)))
	mux.Handle("GET /users/{id}", adminOnly(handlers.UsersGet(userSvc)))
	mux.Handle("POST /users", adminOnly(handlers.UserCreate(userSvc)))
	mux.Handle("DELETE /users/{id}", adminOnly(handlers.UsersDelete(userSvc)))

	// auth
	mux.HandleFunc("POST /auth/signup", handlers.AuthSignup(userSvc))
	mux.HandleFunc("POST /auth/signin", handlers.AuthSignin(userSvc))
	mux.HandleFunc("POST /auth/verify", handlers.AuthVerify(userSvc))
	mux.HandleFunc("POST /auth/resend-verification", handlers.AuthResendVerification(userSvc))
	mux.HandleFunc("POST /auth/reset-password-request", handlers.AuthResetPasswordRequest(userSvc))
	mux.HandleFunc("POST /auth/reset-password-confirm", handlers.AuthResetPasswordConfirm(userSvc))
	mux.Handle("POST /auth/signout", auth(handlers.AuthSignout(userSvc)))

	// version api
	v1 := http.NewServeMux()
	v1.Handle("/v1/", http.StripPrefix("/v1", mux))

	// apply middlewares
	middlewaresStack := middlewares.MiddlewareStack(
		middlewares.Logging,
		middlewares.CommonHeaders,
		middlewares.MaxBodySize,
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
