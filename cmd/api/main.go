package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hreftools/api/internal/emails"
	"github.com/hreftools/api/internal/server"
	"github.com/hreftools/api/internal/store"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/resend/resend-go/v3"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func initTracer() (*sdktrace.TracerProvider, error) {
	exporter, err := otlptrace.New(context.Background(), otlptracehttp.NewClient())
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
	port := os.Getenv("PORT")
	databaseUrl := os.Getenv("DATABASE_URL")
	resendApiKey := os.Getenv("RESEND_API_KEY")

	if port == "" {
		log.Fatal("PORT environment variable is required")
	}
	if databaseUrl == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	pool, err := sql.Open("pgx", databaseUrl)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := pool.Close(); err != nil {
			log.Printf("Failed to close database connection: %v", err)
		}
	}()

	// configure db connection pool
	pool.SetMaxOpenConns(25)
	pool.SetMaxIdleConns(5)
	pool.SetConnMaxLifetime(5 * time.Minute)
	pool.SetConnMaxIdleTime(5 * time.Minute)

	// verify the db connectoin
	if err := pool.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	store := store.New(pool)
	resendClient := resend.NewClient(resendApiKey)
	emailSender := emails.NewResendEmailSender(resendClient)

	tp, err := initTracer()
	if err != nil {
		log.Fatal(err)
	}
	defer tp.Shutdown(context.Background())

	srv := server.New(port, store, emailSender)
	srv.Handler = otelhttp.NewHandler(srv.Handler, "api")

	chServer := make(chan error, 1)

	log.Printf("Starting server on %s", port)
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
