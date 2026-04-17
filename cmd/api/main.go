package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hreftools/api/internal/config"
	"github.com/hreftools/api/internal/db"
	"github.com/hreftools/api/internal/emails"
	"github.com/hreftools/api/internal/postgres"
	"github.com/hreftools/api/internal/resource"
	"github.com/hreftools/api/internal/server"
	"github.com/hreftools/api/internal/user"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/resend/resend-go/v3"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func initTracer(ctx context.Context) (*sdktrace.TracerProvider, error) {
	exporter, err := otlptrace.New(ctx, otlptracehttp.NewClient())
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{}))
	return tp, nil
}

func run(ctx context.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	pool, err := postgres.Connect(cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer func() {
		if err := pool.Close(); err != nil {
			log.Printf("Failed to close database connection: %v", err)
		}
	}()

	queries := db.New(pool)
	userRepo := postgres.NewUserRepository(queries)
	sessionRepo := postgres.NewSessionRepository(queries)
	resourceRepo := postgres.NewResourceRepository(queries)

	resendClient := resend.NewClient(cfg.ResendAPIKey)
	emailSender := emails.NewResendEmailSender(resendClient)

	userSvc := user.NewService(userRepo, sessionRepo, emailSender)
	resourceSvc := resource.NewService(resourceRepo)

	tp, err := initTracer(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize tracer: %w", err)
	}
	defer tp.Shutdown(context.Background())

	srv := server.New(cfg.Port, userSvc, resourceSvc)
	srv.Handler = otelhttp.NewHandler(srv.Handler, "api")

	chServer := make(chan error, 1)

	log.Printf("Starting server on %s", cfg.Port)
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
		log.Printf("Server shutting down due to signal: %v", context.Cause(ctxSignal))
	case err := <-chServer:
		log.Printf("Server error: %v", err)
		return err
	}

	ctxTimeout, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()

	if err := srv.Shutdown(ctxTimeout); err != nil {
		log.Printf("Server shutdown failed: %v", err)

		if closeErr := srv.Close(); closeErr != nil {
			log.Printf("Server close failed: %v", closeErr)
			return errors.Join(err, closeErr)
		}

		return err
	}

	log.Println("Server exited gracefully")
	return nil
}

func main() {
	ctx := context.Background()

	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}
