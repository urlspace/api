package server

import (
	"net/http"
	"time"

	"github.com/hreftools/api/internal/resource"
	"github.com/hreftools/api/internal/user"
)

func New(port string, userSvc *user.Service, resourceSvc *resource.Service) *http.Server {
	// routes
	mux := http.NewServeMux()

	// not found
	mux.HandleFunc("/", handleNotFound)

	// status
	mux.HandleFunc("GET /status", handleStatus)

	auth := authMiddleware(userSvc)
	adminOnly := middlewareStack(auth, adminMiddleware(userSvc))

	// me (authenticated)
	mux.Handle("GET /me", auth(handleMeGet(userSvc)))

	// resources (protected)
	mux.Handle("GET /resources", auth(handleResourcesList(resourceSvc)))
	mux.Handle("GET /resources/{id}", auth(handleResourcesGet(resourceSvc)))
	mux.Handle("POST /resources", auth(handleResourcesCreate(resourceSvc)))
	mux.Handle("PUT /resources/{id}", auth(handleResourcesUpdate(resourceSvc)))
	mux.Handle("DELETE /resources/{id}", auth(handleResourcesDelete(resourceSvc)))

	// users (admin only)
	mux.Handle("GET /users", adminOnly(handleUsersList(userSvc)))
	mux.Handle("GET /users/{id}", adminOnly(handleUsersGet(userSvc)))
	mux.Handle("POST /users", adminOnly(handleUsersCreate(userSvc)))
	mux.Handle("DELETE /users/{id}", adminOnly(handleUsersDelete(userSvc)))

	// auth
	mux.HandleFunc("POST /auth/signup", handleAuthSignup(userSvc))
	mux.HandleFunc("POST /auth/signin", handleAuthSignin(userSvc))
	mux.HandleFunc("POST /auth/verify", handleAuthVerify(userSvc))
	mux.HandleFunc("POST /auth/resend-verification", handleAuthResendVerification(userSvc))
	mux.HandleFunc("POST /auth/reset-password-request", handleAuthResetPasswordRequest(userSvc))
	mux.HandleFunc("POST /auth/reset-password-confirm", handleAuthResetPasswordConfirm(userSvc))
	mux.Handle("POST /auth/signout", auth(handleAuthSignout(userSvc)))

	// version api
	v1 := http.NewServeMux()
	v1.Handle("/v1/", http.StripPrefix("/v1", mux))

	// apply middlewares
	stack := middlewareStack(
		loggingMiddleware,
		commonHeadersMiddleware,
		maxBodySizeMiddleware,
	)

	return &http.Server{
		Addr:              ":" + port,
		Handler:           stack(v1),
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
}
