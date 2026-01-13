package server

import (
	"net/http"
	"os"
	"time"

	"github.com/jumplist/api/internal/handlers"
	"github.com/jumplist/api/internal/middlewares"
	"github.com/jumplist/api/internal/store"
	"github.com/resend/resend-go/v3"
)

func New(store *store.Store, resendClient *resend.Client) *http.Server {
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
	mux.HandleFunc("GET /resources", handlers.ResourcesList(store))
	mux.HandleFunc("GET /resources/{id}", handlers.ResourcesGet(store))
	mux.HandleFunc("POST /resources", handlers.ResourcesCreate(store))
	mux.HandleFunc("PUT /resources/{id}", handlers.ResourcesUpdate(store))
	mux.HandleFunc("DELETE /resources/{id}", handlers.ResourcesDelete(store))

	// users
	// this is for admins only, middleware to come later here
	mux.HandleFunc("GET /users", handlers.UsersList(store))
	mux.HandleFunc("GET /users/{id}", handlers.UsersGet(store))
	mux.HandleFunc("POST /users", handlers.UserCreate(store))
	mux.HandleFunc("DELETE /users/{id}", handlers.UsersDelete(store))

	// auth
	mux.HandleFunc("POST /auth/signup", handlers.AuthSignup(store, resendClient))
	// mux.HandleFunc("POST /auth/signin", handlers.xxx(store))
	// mux.HandleFunc("POST /auth/signout", handlers.xxx(store))
	mux.HandleFunc("POST /auth/verify", handlers.AuthVerify(store))
	// TODO: implement this one to trigger resending a new verification email in case the old token exired
	// mux.HandleFunc("POST /auth/verify-again", handlers.AuthVerifyAgain(store))

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
