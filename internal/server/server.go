package server

import (
	"net/http"
	"time"

	"github.com/urlspace/api/internal/collection"
	"github.com/urlspace/api/internal/tag"
	"github.com/urlspace/api/internal/uow"
	"github.com/urlspace/api/internal/user"
)

func New(port string, userSvc *user.Service, tagSvc *tag.Service, collectionSvc *collection.Service, uowSvc *uow.Service) *http.Server {
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

	// links (protected)
	mux.Handle("GET /links", sessionOrToken(handleLinksList(uowSvc)))
	mux.Handle("GET /links/{id}", sessionOrToken(handleLinksGet(uowSvc)))
	mux.Handle("POST /links", sessionOrToken(handleLinksCreate(uowSvc)))
	mux.Handle("PUT /links/{id}", sessionOrToken(handleLinksUpdate(uowSvc)))
	mux.Handle("DELETE /links/{id}", sessionOrToken(handleLinksDelete(uowSvc)))

	// tags (protected)
	mux.Handle("GET /tags", sessionOrToken(handleTagsList(tagSvc)))
	mux.Handle("PUT /tags/{id}", sessionOrToken(handleTagsUpdate(tagSvc)))
	mux.Handle("DELETE /tags/{id}", sessionOrToken(handleTagsDelete(tagSvc)))

	// collections (protected)
	mux.Handle("GET /collections", sessionOrToken(handleCollectionsList(collectionSvc)))
	mux.Handle("GET /collections/{id}", sessionOrToken(handleCollectionsGet(collectionSvc)))
	mux.Handle("POST /collections", sessionOrToken(handleCollectionsCreate(collectionSvc)))
	mux.Handle("PUT /collections/{id}", sessionOrToken(handleCollectionsUpdate(collectionSvc)))
	mux.Handle("DELETE /collections/{id}", sessionOrToken(handleCollectionsDelete(collectionSvc)))

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
