package server

import (
	"net/http"
	"time"

	"github.com/urlspace/api/internal/collection"
	"github.com/urlspace/api/internal/tag"
	"github.com/urlspace/api/internal/uow"
	"github.com/urlspace/api/internal/user"
)

func New(port string, appURL string, userSvc *user.Service, tagSvc *tag.Service, collectionSvc *collection.Service, uowSvc *uow.Service) *http.Server {
	// routes
	mux := http.NewServeMux()

	sessionAuth := authMiddleware(userSvc, AuthConfig{UseSession: true, UseToken: false})
	sessionOrTokenAuth := authMiddleware(userSvc, AuthConfig{UseSession: true, UseToken: true})
	sessionOrTokenAuthWithPermissions := middlewareStack(sessionOrTokenAuth, userPermissionsMiddleware(userSvc))

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
	mux.Handle("POST /auth/signout", sessionAuth(handleAuthSignout(userSvc)))

	// me (authenticated)
	mux.Handle("GET /me", sessionOrTokenAuth(handleMeGet(userSvc)))
	mux.Handle("POST /me/update-display-name", sessionAuth(handleMeUpdateDisplayName(userSvc)))
	mux.Handle("POST /me/update-username", sessionAuth(handleMeUpdateUsername(userSvc)))
	mux.Handle("POST /me/update-email", sessionAuth(handleMeUpdateEmail(userSvc)))
	mux.Handle("POST /me/update-email-confirm", sessionAuth(handleMeUpdateEmailConfirm(userSvc)))
	mux.Handle("POST /me/update-password", sessionAuth(handleMeUpdatePassword(userSvc)))
	mux.Handle("DELETE /me", sessionAuth(handleMeDelete(userSvc)))

	// links (protected)
	mux.Handle("GET /links", sessionOrTokenAuth(handleLinksList(uowSvc)))
	mux.Handle("GET /links/{id}", sessionOrTokenAuth(handleLinksGet(uowSvc)))
	mux.Handle("POST /links", sessionOrTokenAuth(handleLinksCreate(uowSvc)))
	mux.Handle("PUT /links/{id}", sessionOrTokenAuth(handleLinksUpdate(uowSvc)))
	mux.Handle("DELETE /links/{id}", sessionOrTokenAuth(handleLinksDelete(uowSvc)))

	// tags (protected)
	mux.Handle("GET /tags", sessionOrTokenAuth(handleTagsList(tagSvc)))
	mux.Handle("PUT /tags/{id}", sessionOrTokenAuth(handleTagsUpdate(tagSvc)))
	mux.Handle("DELETE /tags/{id}", sessionOrTokenAuth(handleTagsDelete(tagSvc)))

	// collections (protected)
	mux.Handle("GET /collections", sessionOrTokenAuth(handleCollectionsList(collectionSvc)))
	mux.Handle("GET /collections/{id}", sessionOrTokenAuth(handleCollectionsGet(collectionSvc)))
	mux.Handle("POST /collections", sessionOrTokenAuthWithPermissions(handleCollectionsCreate(collectionSvc)))
	mux.Handle("PUT /collections/{id}", sessionOrTokenAuthWithPermissions(handleCollectionsUpdate(collectionSvc)))
	mux.Handle("DELETE /collections/{id}", sessionOrTokenAuth(handleCollectionsDelete(collectionSvc)))

	// sessions (session-only — session management requires an active session)
	mux.Handle("GET /sessions", sessionAuth(handleSessionsList(userSvc)))
	mux.Handle("DELETE /sessions/{id}", sessionAuth(handleSessionsDelete(userSvc)))
	mux.Handle("DELETE /sessions", sessionAuth(handleSessionsDeleteAll(userSvc)))

	// tokens (session-only — token management requires an active session)
	mux.Handle("POST /tokens", sessionAuth(handleTokensCreate(userSvc)))
	mux.Handle("GET /tokens", sessionAuth(handleTokensList(userSvc)))
	mux.Handle("GET /tokens/{id}", sessionAuth(handleTokensGet(userSvc)))
	mux.Handle("DELETE /tokens/{id}", sessionAuth(handleTokensDelete(userSvc)))
	mux.Handle("DELETE /tokens", sessionAuth(handleTokensDeleteAll(userSvc)))

	// users (admin only)
	mux.Handle("GET /admin/users", sessionOrTokenAuthWithPermissions(handleUsersList(userSvc)))
	mux.Handle("GET /admin/users/{id}", sessionOrTokenAuthWithPermissions(handleUsersGet(userSvc)))
	mux.Handle("POST /admin/users", sessionOrTokenAuthWithPermissions(handleUsersCreate(userSvc)))
	mux.Handle("DELETE /admin/users/{id}", sessionOrTokenAuthWithPermissions(handleUsersDelete(userSvc)))

	// version api
	v1 := http.NewServeMux()
	v1.Handle("/v1/", http.StripPrefix("/v1", mux))

	// apply middlewares
	stack := middlewareStack(
		loggingMiddleware,
		commonHeadersMiddleware(appURL),
		recoveryMiddleware,
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
