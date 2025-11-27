package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

type NotFoundResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func NotFound(w http.ResponseWriter, r *http.Request) {

	response := &NotFoundResponse{
		Status: "error",
		Data:   "endpoint not found",
	}

	w.WriteHeader(http.StatusNotFound)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding error response: %v", err)
	}
}
