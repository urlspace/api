package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/resend/resend-go/v3"
	"github.com/urlspace/api/internal/collection"
	"github.com/urlspace/api/internal/config"
	"github.com/urlspace/api/internal/db"
	"github.com/urlspace/api/internal/emails"
	"github.com/urlspace/api/internal/postgres"
	"github.com/urlspace/api/internal/server"
	"github.com/urlspace/api/internal/tag"
	"github.com/urlspace/api/internal/telemetry"
	"github.com/urlspace/api/internal/uow"
	"github.com/urlspace/api/internal/user"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func run(ctx context.Context) error {
	shutdownTelemetry, err := telemetry.Setup(ctx)
	if err != nil {
		return err
	}
	defer shutdownTelemetry(context.Background())

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	slog.Info("running database migrations")
	if err := postgres.Migrate(cfg.DatabaseURL); err != nil {
		return err
	}
	slog.Info("database migrations complete")

	pool, err := postgres.Connect(cfg.DatabaseURL, telemetry.NewPgxTracer())
	if err != nil {
		return err
	}
	defer pool.Close()

	queries := db.New(pool)
	userRepo := postgres.NewUserRepository(queries)
	sessionRepo := postgres.NewSessionRepository(queries)
	tokenRepo := postgres.NewTokenRepository(queries)
	linkRepo := postgres.NewLinkRepository(queries)
	tagRepo := postgres.NewTagRepository(queries)
	collectionRepo := postgres.NewCollectionRepository(queries)

	unitOfWork := postgres.NewUnitOfWork(pool)

	resendHTTPClient := &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
	resendClient := resend.NewCustomClient(resendHTTPClient, cfg.ResendAPIKey)
	emailSender := emails.NewResendEmailSender(resendClient)

	userSvc := user.NewService(userRepo, sessionRepo, tokenRepo, emailSender, cfg.AppURL, cfg.AdminEmail)
	tagSvc := tag.NewService(tagRepo)
	collectionSvc := collection.NewService(collectionRepo)
	uowSvc := uow.NewService(uow.Repositories{
		Links:       linkRepo,
		Tags:        tagRepo,
		Collections: collectionRepo,
	}, unitOfWork)

	srv := server.New(cfg.Port, cfg.AppURL, userSvc, tagSvc, collectionSvc, uowSvc)
	srv.Handler = otelhttp.NewHandler(srv.Handler, "api")

	chServer := make(chan error, 1)

	slog.Info("starting server", slog.String("port", cfg.Port))
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			chServer <- err
		}
		close(chServer)
	}()

	ctxSignal, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case <-ctxSignal.Done():
		slog.Info("shutting down server", slog.String("signal", context.Cause(ctxSignal).Error()))
	case err := <-chServer:
		slog.Error("server error", slog.String("error", err.Error()))
		return err
	}

	ctxTimeout, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()

	if err := srv.Shutdown(ctxTimeout); err != nil {
		slog.Error("server shutdown failed", slog.String("error", err.Error()))

		if closeErr := srv.Close(); closeErr != nil {
			slog.Error("server close failed", slog.String("error", closeErr.Error()))
			return errors.Join(err, closeErr)
		}

		return err
	}

	slog.Info("server exited gracefully")
	return nil
}

func main() {
	ctx := context.Background()

	if err := run(ctx); err != nil {
		slog.Error("fatal error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
