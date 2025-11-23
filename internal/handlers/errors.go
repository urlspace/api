package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type ErrorResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func HandleDbError(w http.ResponseWriter, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		writeJSONError(w, http.StatusNotFound, "resource not found")
		return
	}

	if errors.Is(err, context.DeadlineExceeded) {
		writeJSONError(w, http.StatusRequestTimeout, "request timeout")
		return
	}

	if errors.Is(err, context.Canceled) {
		writeJSONError(w, 499, "request cancelled")
		return
	}

	log.Printf("Database error: %v", err)
	writeJSONError(w, http.StatusInternalServerError, "internal server error")
}

func HandleClientError(w http.ResponseWriter, err error, message string) {
	log.Printf("Client error: %v", err)
	writeJSONError(w, http.StatusBadRequest, message)
}

func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)

	response := &ErrorResponse{
		Status: "error",
		Data:   message,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding error response: %v", err)
	}
}
