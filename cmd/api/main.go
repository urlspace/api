package main

import (
	"github.com/zapi-sh/api/internal/server"
	"log"
	"net/http"
)

func main() {
	s := server.New()

	log.Printf("Starting server on %s", s.Addr)

	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}
