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

	sessionOnly := authMiddleware(userSvc, AuthConfig{UseSession: true, UseToken: false})
	sessionOrToken := authMiddleware(userSvc, AuthConfig{UseSession: true, UseToken: true})
	adminOnly := middlewareStack(sessionOrToken, adminMiddleware(userSvc))

	// not found
	mux.HandleFunc("/", handleNotFound)

	// status
	mux.HandleFunc("GET /status", handleStatus)

	// auth
	mux.HandleFunc("POST /auth/signup", handleAuthSignup(userSvc))
	mux.HandleFunc("POST /auth/signin", handleAuthSignin(userSvc))
	mux.HandleFunc("POST /auth/verify", handleAuthVerify(userSvc))
	mux.HandleFunc("POST /auth/resend-verification", handleAuthResendVerification(userSvc))
	mux.HandleFunc("POST /auth/reset-password-request", handleAuthResetPasswordRequest(userSvc))
	mux.HandleFunc("POST /auth/reset-password-confirm", handleAuthResetPasswordConfirm(userSvc))
	mux.Handle("POST /auth/signout", sessionOnly(handleAuthSignout(userSvc)))
	mux.Handle("DELETE /auth", sessionOnly(handleAuthDelete(userSvc)))

	// me (authenticated)
	mux.Handle("GET /me", sessionOrToken(handleMeGet(userSvc)))

	// resources (protected)
	mux.Handle("GET /resources", sessionOrToken(handleResourcesList(resourceSvc)))
	mux.Handle("GET /resources/{id}", sessionOrToken(handleResourcesGet(resourceSvc)))
	mux.Handle("POST /resources", sessionOrToken(handleResourcesCreate(resourceSvc)))
	mux.Handle("PUT /resources/{id}", sessionOrToken(handleResourcesUpdate(resourceSvc)))
	mux.Handle("DELETE /resources/{id}", sessionOrToken(handleResourcesDelete(resourceSvc)))

	// tokens (session-only — token management requires an active session)
	mux.Handle("POST /tokens", sessionOnly(handleTokensCreate(userSvc)))
	mux.Handle("GET /tokens", sessionOnly(handleTokensList(userSvc)))
	mux.Handle("GET /tokens/{id}", sessionOnly(handleTokensGet(userSvc)))
	mux.Handle("DELETE /tokens/{id}", sessionOnly(handleTokensDelete(userSvc)))
	mux.Handle("DELETE /tokens", sessionOnly(handleTokensDeleteAll(userSvc)))

	// users (admin only)
	mux.Handle("GET /admin/users", adminOnly(handleUsersList(userSvc)))
	mux.Handle("GET /admin/users/{id}", adminOnly(handleUsersGet(userSvc)))
	mux.Handle("POST /admin/users", adminOnly(handleUsersCreate(userSvc)))
	mux.Handle("DELETE /admin/users/{id}", adminOnly(handleUsersDelete(userSvc)))

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
