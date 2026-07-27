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
)

func run(ctx context.Context) error {
	// otelc auto instrumentation configures tracing and metrics.
	// logging is the only thing that needs a little bit of setup to setup
	// multiHandler that propagates the logs to the otel endpoint and to stdout
	telemetry.InitLogging()

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

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

	resendClient := resend.NewClient(cfg.ResendAPIKey)
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
