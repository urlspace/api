package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/zapi-sh/api/internal/server"
)

func main() {
	s := server.New()

	go func() {
		log.Printf("Starting server on %s", s.Addr)
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	ctxSignal, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctxSignal.Done()

	log.Println("Server shutting down.")
	ctxTimeout, stop := context.WithTimeout(context.Background(), 10*time.Second)
	defer stop()
	if err := s.Shutdown(ctxTimeout); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
	log.Println("Server closed.")

}
